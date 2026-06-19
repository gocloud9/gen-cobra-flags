// Package generator implements the gen-cobra-flags code generator.
//
// It parses Go source annotated with +cobra:* markers and produces Cobra
// flag-binding boilerplate. The public entry point is Generate.
package generator

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/gocloud9/gen-tool/pkg/generate"
	"github.com/gocloud9/gen-tool/pkg/parse"
)

//go:embed templates/*
var templateFS embed.FS

// Options configures a generation run.
type Options struct {
	// InputDir is the directory to parse for annotated structs.
	InputDir string

	// OutputDir is the directory where generated files are written.
	OutputDir string

	// Package is the package name used in the generated files' package clause.
	Package string

	// Struct, when non-empty, restricts generation to the named struct.
	// When empty, every annotated struct in the input is processed.
	Struct string

	// SamePackageAsSource indicates the generated code lives in the same
	// package as the source structs, so field types are referenced without
	// the external package qualifier.
	//
	// Callers normally leave this false: Generate auto-detects the
	// same-package case by comparing the resolved InputDir and OutputDir. When
	// the generated files are written into the source directory, they share
	// the source package, so no source-package import or qualifier is emitted.
	// Setting it true forces same-package behavior regardless of directories.
	SamePackageAsSource bool

	// SourceImport is the import path of the package containing the source
	// structs (e.g. "github.com/acme/app/api"). Required when
	// SamePackageAsSource is false, so the generated To<Struct> methods can
	// reference the original types. Ignored when SamePackageAsSource is true.
	SourceImport string
}

// Generate runs the generator with the given options, writing generated files
// into opts.OutputDir.
func Generate(opts Options) error {
	if opts.InputDir == "" {
		return fmt.Errorf("input directory is required")
	}
	if opts.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}
	if opts.Package == "" {
		return fmt.Errorf("output package name is required")
	}

	// Auto-detect the same-package case: when the generated files are written
	// into the source directory, they share the source package, so no
	// source-package import or qualifier must be emitted. An explicitly set
	// SamePackageAsSource still forces same-package behavior.
	if same, err := sameDir(opts.InputDir, opts.OutputDir); err != nil {
		return fmt.Errorf("comparing input and output directories: %w", err)
	} else if same {
		opts.SamePackageAsSource = true
	}

	// For the cross-package case, the generated To<Struct> methods reference
	// the source types with a package qualifier, so the generated file must
	// import the source package. When the caller did not supply an explicit
	// -source-import, derive it from the input directory's Go module.
	if !opts.SamePackageAsSource && opts.SourceImport == "" {
		derived, err := deriveSourceImport(opts.InputDir)
		if err != nil {
			return fmt.Errorf("deriving source import path for %q (pass -source-import to set it explicitly): %w", opts.InputDir, err)
		}
		opts.SourceImport = derived
	}

	g := &gen{opts: opts}

	parser := parse.Parser{}
	results, err := parser.ParseDirectory(parse.Options{
		Path: opts.InputDir,
		SkipFilesWithContentsRegex: []*regexp.Regexp{
			regexp.MustCompile("Generated Code with gen-cobra-flags - Do Not Edit"),
		},
	})
	if err != nil {
		return fmt.Errorf("parsing input directory %q: %w", opts.InputDir, err)
	}

	g.results = results

	// Selection is driven by the +cobra:enabled marker (replacing the former
	// -struct CLI flag). Keep only enabled structs plus any child structs
	// pulled in by +cobra:config:child / +cobra:subcommand fields, and
	// synthesize the markers those children need for generation.
	g.selectStructs(results)

	if opts.Struct != "" {
		filterToStruct(results, opts.Struct)
	}

	flagsDest := filepath.Join(opts.OutputDir, "{{.Struct.Name | toSnakeCase }}_gen.go")
	sharedDest := filepath.Join(opts.OutputDir, "shared.go")

	err = generate.ExecuteWithCustom(results, generate.OptionsWithCustom[customInput]{
		CustomInput: customInput{DestinationPackage: opts.Package},
		EmdedFS:     []embed.FS{templateFS},
		Files: generate.Files{
			{
				TemplatePath:    "templates/flags.go.tmpl",
				DestinationPath: flagsDest,
				Type:            generate.PerStruct,
				FormatSource:    true,
			},
			{
				TemplatePath:    "templates/shared.go.tmpl",
				DestinationPath: sharedDest,
				Type:            generate.Global,
				FormatSource:    true,
			},
		},
		TemplateFuncMap: g.funcMap(),
	})
	if err != nil {
		return fmt.Errorf("generating output: %w", err)
	}

	return nil
}

// customInput is passed to templates as .Custom.
type customInput struct {
	DestinationPackage string
}

// filterToStruct removes all structs except the one named from the parse
// results, so generation targets only that struct.
func filterToStruct(results *parse.Results, name string) {
	for i := range results.Packages {
		for key, s := range results.Packages[i].Structs {
			if s.Name != name {
				delete(results.Packages[i].Structs, key)
			}
		}
	}
}

// selectStructs reduces the parse results to the set of structs that should be
// generated. Selection is driven by the +cobra:enabled struct marker: every
// enabled struct is kept, along with the child structs referenced by its
// +cobra:config:child / +cobra:subcommand fields. Child structs are mutated in
// place so the existing per-struct template can emit a full child config: they
// gain a synthesized +cobra:flag (so a config aggregate flag exists) and their
// +cobra:required fields gain auto-derived +cobra:flag / +cobra:usage markers.
//
// For backward compatibility, if no struct in the package declares
// +cobra:enabled, selection falls back to the prior behavior (all structs,
// optionally narrowed by the -struct CLI flag).
func (g *gen) selectStructs(results *parse.Results) {
	if !anyEnabled(results) {
		return
	}

	for i := range results.Packages {
		pkg := results.Packages[i]

		keep := map[string]bool{}
		for name, s := range pkg.Structs {
			if g.strct(s).enabled() {
				keep[name] = true
			}
		}

		// Pull in child structs referenced by enabled structs and prepare them
		// for generation.
		for name := range keep {
			g.collectChildren(pkg, pkg.Structs[name], keep)
		}

		for name := range pkg.Structs {
			if !keep[name] {
				delete(pkg.Structs, name)
			}
		}
	}
}

// anyEnabled reports whether any parsed struct carries +cobra:enabled.
func anyEnabled(results *parse.Results) bool {
	for i := range results.Packages {
		for _, s := range results.Packages[i].Structs {
			if (&Struct{StructInfo: s}).enabled() {
				return true
			}
		}
	}
	return false
}

// collectChildren walks the +cobra:config:child / +cobra:subcommand fields of s
// and marks their element structs for generation, recursively. Each child is
// prepared (markers synthesized) so the existing per-struct template emits a
// valid child config.
func (g *gen) collectChildren(pkg *parse.PackageInfo, s *parse.StructInfo, keep map[string]bool) {
	for k := range s.Fields {
		field := g.field(s.Fields[k])
		if !field.isConfigChild() && !field.isSubcommand() {
			continue
		}
		childName := field.childTypeName()
		child, ok := pkg.Structs[childName]
		if !ok {
			// Child type is not declared in this package; nothing to generate.
			continue
		}
		if keep[childName] {
			continue
		}
		keep[childName] = true
		prepareChildStruct(child)
		// Recurse so grandchildren marked config:child are pulled in too.
		g.collectChildren(pkg, child, keep)
	}
}

// prepareChildStruct synthesizes the markers a child struct needs so the
// per-struct template generates a complete child config. The child gains an
// aggregate +cobra:flag (derived from its name) when it has none, and each
// +cobra:required field gains an auto-derived +cobra:flag and +cobra:usage when
// it lacks them. Existing explicit markers are never overwritten.
func prepareChildStruct(s *parse.StructInfo) {
	if s.Markers == nil {
		s.Markers = map[string]string{}
	}
	if _, ok := s.Markers["+cobra:flag"]; !ok {
		s.Markers["+cobra:flag"] = toKebabCase(s.Name) + "-config"
	}
	if _, ok := s.Markers["+cobra:usage"]; !ok {
		s.Markers["+cobra:usage"] = "Aggregate JSON/YAML config for " + s.Name + "."
	}

	for k := range s.Fields {
		f := s.Fields[k]
		if f.Markers == nil {
			f.Markers = map[string]string{}
		}
		// Only +cobra:required child fields are exposed as individual flags.
		if !(&Field{FieldInfo: f}).required() {
			continue
		}
		if !isExported(f.Name) {
			continue
		}
		if _, ok := f.Markers["+cobra:flag"]; !ok {
			f.Markers["+cobra:flag"] = toKebabCase(f.Name)
		}
		if _, ok := f.Markers["+cobra:usage"]; !ok {
			f.Markers["+cobra:usage"] = f.Name + " for " + s.Name + "."
		}
	}
}

// gen holds per-run state and the template helper methods.
type gen struct {
	opts    Options
	results *parse.Results
}

func (g *gen) funcMap() template.FuncMap {
	return template.FuncMap{
		"toRegisterMethod": func(f *AdaptorInfo) string {
			return fmt.Sprintf("Register%s%s", strings.ToUpper(f.Name)[:1], f.Name[1:])
		},
		"fieldToFlagMethod": func(f *parse.FieldInfo) string {
			field := g.field(f)
			if field.TypeName == "time.Time" {
				return fmt.Sprintf("%sP(%q, %q, %s, []string{time.RFC3339}, %q)", field.flagType(), field.flag(), field.short(), field.flagDefault(), field.flagUsage())
			}
			return fmt.Sprintf("%sP(%q, %q, %s, %q)", field.flagType(), field.flag(), field.short(), field.flagDefault(), field.flagUsage())
		},
		"fieldToFlagGetMethod": func(f *parse.FieldInfo) string {
			field := g.field(f)
			return fmt.Sprintf("Get%s(%q)", field.flagType(), field.flag())
		},
		"getCobraTags": func(field *parse.FieldInfo) string {
			tags := []string{}
			if v := field.Markers["+cobra:json"]; v != "" {
				tags = append(tags, fmt.Sprintf("json:%q", v))
			}
			if v := field.Markers["+cobra:yaml"]; v != "" {
				tags = append(tags, fmt.Sprintf("yaml:%q", v))
			}
			if v := field.Markers["+cobra:customTags"]; v != "" {
				tags = append(tags, v)
			}
			if len(tags) == 0 {
				return ""
			}
			return fmt.Sprintf(" `%s`", strings.Join(tags, " "))
		},
		"onlyCobraFlags": func(in map[string]*parse.FieldInfo) map[string]*parse.FieldInfo {
			out := map[string]*parse.FieldInfo{}
			for i := range in {
				if g.field(in[i]).hasFlag() {
					out[i] = in[i]
				}
			}
			return out
		},
		"getAdaptors":           g.getAdaptors,
		"sharedImports":         g.sharedImports,
		"onlyCobraOptions":      onlyCobraOptions,
		"getImports":            g.getImports,
		"asConfigAdaptorName":   func(info *parse.FieldInfo) string { return g.field(info).configAdaptor() },
		"asFlagAdaptorName":     func(info *parse.FieldInfo) string { return g.field(info).flagAdaptor() },
		"fieldToConfigTypeName": func(f *parse.FieldInfo) string { return g.field(f).configType() },
		// fieldIsConfigChild reports whether the field carries
		// +cobra:config:child, so its parent-config representation is the
		// generated <Child>Config type and To<Struct> must convert each
		// element back to the domain type via To<Child>.
		"fieldIsConfigChild": func(f *parse.FieldInfo) bool { return g.field(f).isConfigChild() },
		// fieldDomainType returns the source domain type of the field (e.g.
		// "[]*types.CreateVpcPeer"), used as the target type in the
		// config:child conversion in To<Struct>.
		"fieldDomainType": func(f *parse.FieldInfo) string { return g.field(f).domainType() },
		// fieldToChildMethod returns the conversion method name on the child
		// config that yields the domain type, e.g. "ToCreateVpcPeer".
		"fieldToChildMethod": func(f *parse.FieldInfo) string { return g.field(f).toChildMethod() },
		// fieldConfigChildIsSlice reports whether a config:child field is a
		// slice, so To<Struct> emits an element-wise conversion loop rather
		// than a single conversion.
		"fieldConfigChildIsSlice": func(f *parse.FieldInfo) bool { return g.field(f).configChildIsSlice() },
		"structToFlagAdaptorName": func(s *parse.StructInfo) string {
			return g.strct(s).flagAdaptor()
		},
		"structToFlagMethod": func(s *parse.StructInfo) string {
			st := g.strct(s)
			return fmt.Sprintf("StringP(%q, %q, %s, %q)", st.flag(), st.short(), st.defaultValue(), st.usage())
		},
		"structToFlagGetMethod": func(s *parse.StructInfo) string {
			return fmt.Sprintf("GetString(%q)", g.strct(s).flag())
		},
		"needsConfigAdaptor": func(f *parse.FieldInfo) bool {
			field := g.field(f)
			// config:child fields are handled by a dedicated To<Struct>
			// branch that converts each element via To<Child>; they must not
			// fall into the generic config-adaptor path.
			if field.isConfigChild() {
				return false
			}
			if _, ok := field.Markers["+cobra:config:adaptor"]; ok {
				return true
			}
			return !field.configIsSameType()
		},
		"needsFlagAdaptor": func(f *parse.FieldInfo) bool {
			field := g.field(f)
			if _, ok := field.Markers["+cobra:flag:adaptor"]; ok {
				return true
			}
			return field.configType() != CobraToGoType(field.flagType())
		},
		"toSnakeCase": toSnakeCase,
		"sortedFields": func(in map[string]*parse.FieldInfo) []*parse.FieldInfo {
			return sortedFields(in)
		},
		"exportedFields": func(in map[string]*parse.FieldInfo) []*parse.FieldInfo {
			return sortedFields(onlyExported(in))
		},
		"requiredFields": func(in map[string]*parse.FieldInfo) []*parse.FieldInfo {
			return sortedFields(g.onlyRequired(in))
		},
		"hasRequired": func(in map[string]*parse.FieldInfo) bool {
			return len(g.onlyRequired(in)) > 0
		},
		"fieldIsRequired": func(f *parse.FieldInfo) bool { return g.field(f).required() },
		"requiredImports": g.requiredImports,
		"fieldZeroValue": func(f *parse.FieldInfo) string {
			return g.field(f).configZero()
		},
		"fieldDefaultValue": func(f *parse.FieldInfo) string {
			return g.field(f).flagDefault()
		},
		"fieldConfigIsPrimitive": g.fieldConfigIsPrimitive,
		"fieldTargetIsPrimitive": func(f *parse.FieldInfo) bool {
			return isPrimitive(g.field(f).TypeName)
		},
		"getFlagName": func(f *parse.FieldInfo) string { return g.field(f).flag() },
		// samePackage reports whether generation targets the same package as
		// the source struct. When true, the To<Struct> method must reference
		// the target type without a package qualifier.
		"samePackage": func() bool { return g.opts.SamePackageAsSource },
		// hasOptions reports whether the struct declares any +cobra:option
		// fields, so the To<Struct> method can avoid emitting an empty,
		// unused-variable range loop over opts.
		"hasOptions": func(fields map[string]*parse.FieldInfo) bool {
			return len(onlyCobraOptions(fields)) > 0
		},
		// needsReflect reports whether any field comparison requires
		// reflect.DeepEqual, so the import can be omitted when unused.
		"needsReflect": g.needsReflect,
		// subcommands returns the +cobra:subcommand specs declared on the
		// struct, each resolving the singular child resource plus the parent's
		// required fields.
		"subcommands": g.subcommands,
		// hasSubcommands reports whether the struct declares any
		// +cobra:subcommand fields.
		"hasSubcommands": func(s *parse.StructInfo) bool {
			return len(g.subcommands(s)) > 0
		},
		// childFlagFields returns the child struct fields exposed as individual
		// flags on a subcommand: those marked +cobra:required (auto-derived into
		// flags during selection).
		"childFlagFields": func(s *parse.StructInfo) []*parse.FieldInfo {
			return sortedFields(g.onlyRequired(s.Fields))
		},
		// subcommandName builds the singular subcommand identifier for a field,
		// e.g. parent "UpdateVpcRequest" + field "AddPeers" -> "AddPeer".
		"subcommandName": subcommandName,
		// subcommandConfigStructName returns the composed config struct name,
		// e.g. "UpdateVpcRequestCreateVpcPeerConfig".
		"subcommandConfigStructName": func(sc SubcommandInfo) string {
			return sc.configStructName()
		},
		// subcommandRequiredFieldsStructName returns the per-parent hoisted
		// required-fields struct name, e.g. "UpdateVpcRequestChildRequiredFields".
		"subcommandRequiredFieldsStructName": func(sc SubcommandInfo) string {
			return sc.requiredFieldsStructName()
		},
		// subcommandChildConfigName returns the child's own config struct name,
		// e.g. "CreateVpcPeerConfig".
		"subcommandChildConfigName": func(sc SubcommandInfo) string {
			return sc.childConfigName()
		},
		// subcommandAggregateFlag returns the composed-config aggregate flag
		// name for the subcommand (from +cobra:subcommand:config:flag, or a
		// "<child>-config" fallback).
		"subcommandAggregateFlag": func(sc SubcommandInfo) string {
			return sc.aggregateFlag()
		},
		// subcommandAggregateShort returns the short flag for the composed
		// aggregate flag (from +cobra:subcommand:config:short; may be empty).
		"subcommandAggregateShort": func(sc SubcommandInfo) string {
			return sc.aggregateShort()
		},
		// subcommandAggregateUsage returns the usage string for the composed
		// aggregate flag (from +cobra:subcommand:config:usage, or a fallback).
		"subcommandAggregateUsage": func(sc SubcommandInfo) string {
			return sc.aggregateUsage()
		},
		// subcommandParentTypeRef returns the parent domain type as referenced
		// from the generated package (bare or package-qualified), used as the
		// return type of <P><C>ConfigFromFlags.
		"subcommandParentTypeRef": func(sc SubcommandInfo) string {
			return sc.parentTypeRef(g.opts.SamePackageAsSource, g.pkgName())
		},
		// subcommandChildToMethod returns the child config's conversion method
		// name (e.g. "ToCreateVpcPeer") used in the final conversion.
		"subcommandChildToMethod": func(sc SubcommandInfo) string {
			return sc.childToMethod()
		},
		// subcommandSingularField returns the parent field the subcommand acts
		// on (e.g. "AddPeers"); the final conversion writes the single child
		// into it.
		"subcommandSingularField": func(sc SubcommandInfo) string {
			return sc.singularFieldName()
		},
		// subcommandSingularIsSlice reports whether the parent field is a
		// repeated resource, so the final conversion wraps the child in a
		// one-element slice.
		"subcommandSingularIsSlice": func(sc SubcommandInfo) bool {
			return sc.singularFieldIsSlice()
		},
		// subcommandChildFlagFields returns the child's individual flag fields
		// (its +cobra:required fields) for overlaying onto the composed config.
		"subcommandChildFlagFields": func(sc SubcommandInfo) []*parse.FieldInfo {
			return sortedFields(g.onlyRequired(sc.Child.Fields))
		},
		// subcommandIsScalarSlice reports whether the subcommand operates on a
		// repeated primitive field (e.g. []string) with no struct child, so the
		// template emits the scalar-value variant instead of the child-config
		// variant.
		"subcommandIsScalarSlice": func(sc SubcommandInfo) bool {
			return sc.isScalarSlice()
		},
		// subcommandScalarFieldName returns the singular Go field name used in
		// the composed config for a scalar-slice subcommand (e.g. "RemovePeer").
		"subcommandScalarFieldName": func(sc SubcommandInfo) string {
			return sc.scalarFieldName()
		},
		// subcommandScalarValueFlag returns the flag name for the single scalar
		// value of a scalar-slice subcommand (e.g. "peer-vpc-id").
		"subcommandScalarValueFlag": func(sc SubcommandInfo) string {
			return sc.scalarValueFlag()
		},
		// subcommandScalarValueShort returns the short flag for the scalar value
		// of a scalar-slice subcommand (may be empty).
		"subcommandScalarValueShort": func(sc SubcommandInfo) string {
			return sc.scalarValueShort()
		},
		// subcommandScalarValueUsage returns the usage string for the scalar
		// value of a scalar-slice subcommand.
		"subcommandScalarValueUsage": func(sc SubcommandInfo) string {
			return sc.scalarValueUsage()
		},
	}
}

// needsReflect reports whether generating the given struct emits any
// reflect.DeepEqual call (i.e. it has a non-primitive cobra-flag field).
func (g *gen) needsReflect(s *parse.StructInfo) bool {
	for k := range s.Fields {
		f := s.Fields[k]
		if !g.field(f).hasFlag() {
			continue
		}
		if !g.fieldConfigIsPrimitive(f) {
			return true
		}
	}
	return false
}

func (g *gen) field(f *parse.FieldInfo) *Field {
	return &Field{FieldInfo: f, samePackageAsSource: g.opts.SamePackageAsSource}
}

func (g *gen) strct(s *parse.StructInfo) *Struct {
	return &Struct{StructInfo: s}
}

// pkgName returns the source package name (used to qualify domain types when
// generating into a different package), deterministically selecting the
// package that declares the annotated structs.
func (g *gen) pkgName() string {
	if g.results == nil {
		return ""
	}
	return sourcePackageName(g.results)
}

func toSnakeCase(in string) string {
	return strings.ToLower(regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(in, "${1}_${2}"))
}

// sameDir reports whether two directory paths resolve to the same location. It
// compares the cleaned absolute paths, evaluating symlinks when both paths
// exist so that, e.g., "./" and the fully-qualified source directory are
// recognized as identical. A path that does not yet exist (e.g. an output
// directory created later) falls back to absolute-path comparison.
func sameDir(a, b string) (bool, error) {
	absA, err := filepath.Abs(a)
	if err != nil {
		return false, err
	}
	absB, err := filepath.Abs(b)
	if err != nil {
		return false, err
	}
	if resolved, err := filepath.EvalSymlinks(absA); err == nil {
		absA = resolved
	}
	if resolved, err := filepath.EvalSymlinks(absB); err == nil {
		absB = resolved
	}
	return filepath.Clean(absA) == filepath.Clean(absB), nil
}

// deriveSourceImport computes the Go import path of the package in dir by
// locating the enclosing module's go.mod, reading its module path, and joining
// it with dir's path relative to the module root. For example, a module
// "github.com/acme/app" whose go.mod sits two levels above dir "pkg/types"
// yields "github.com/acme/app/pkg/types". The module root itself yields the
// bare module path.
func deriveSourceImport(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}

	modRoot, modPath, err := findModule(abs)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(modRoot, abs)
	if err != nil {
		return "", err
	}
	if rel == "." || rel == "" {
		return modPath, nil
	}
	// Import paths always use forward slashes regardless of OS.
	return path.Join(modPath, filepath.ToSlash(rel)), nil
}

// findModule walks up from dir until it finds a go.mod, returning the directory
// containing it and the declared module path.
func findModule(dir string) (modRoot, modPath string, err error) {
	for {
		goMod := filepath.Join(dir, "go.mod")
		if data, readErr := os.ReadFile(goMod); readErr == nil {
			mp := modulePath(data)
			if mp == "" {
				return "", "", fmt.Errorf("go.mod at %q has no module directive", goMod)
			}
			return dir, mp, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", fmt.Errorf("no go.mod found in or above %q", dir)
		}
		dir = parent
	}
}

// modulePath extracts the module path from go.mod contents, ignoring comments
// and block-quoted module statements.
func modulePath(goMod []byte) string {
	for _, line := range strings.Split(string(goMod), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "module") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, "module"))
		// Strip a trailing line comment and optional quoting.
		if i := strings.Index(rest, "//"); i >= 0 {
			rest = strings.TrimSpace(rest[:i])
		}
		rest = strings.Trim(rest, "\"`")
		if rest != "" {
			return rest
		}
	}
	return ""
}

// toKebabCase converts an exported Go identifier to a hyphenated flag name,
// e.g. "PeerVpcId" -> "peer-vpc-id".
func toKebabCase(in string) string {
	return strings.ToLower(regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(in, "${1}-${2}"))
}

// subcommandName returns a singular identifier for a repeated subcommand field,
// e.g. "AddPeers" -> "AddPeer", "Addresses" -> "Address". It is used to name
// the generated per-element subcommand helpers.
func subcommandName(in string) string {
	switch {
	case strings.HasSuffix(in, "ies"):
		return strings.TrimSuffix(in, "ies") + "y"
	case strings.HasSuffix(in, "sses"), strings.HasSuffix(in, "xes"),
		strings.HasSuffix(in, "ches"), strings.HasSuffix(in, "shes"):
		return strings.TrimSuffix(in, "es")
	case strings.HasSuffix(in, "s") && !strings.HasSuffix(in, "ss"):
		return strings.TrimSuffix(in, "s")
	default:
		return in
	}
}

func isPrimitive(typeName string) bool {
	switch typeName {
	case "string", "bool",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return true
	default:
		return false
	}
}

func (g *gen) fieldConfigIsPrimitive(f *parse.FieldInfo) bool {
	field := g.field(f)
	if isPrimitive(field.configType()) {
		return true
	}
	if !field.configIsSameType() {
		return false
	}
	return isPrimitive(GetBaseType(field.TypeInfo).TypeName)
}

// onlyRequired returns the subset of fields marked +cobra:required that are
// also exposed as individual cobra flags. Requiredness is validated after the
// aggregate config flag and individual flags are merged, so a value supplied
// through either path satisfies the requirement.
func (g *gen) onlyRequired(in map[string]*parse.FieldInfo) map[string]*parse.FieldInfo {
	out := map[string]*parse.FieldInfo{}
	for i := range in {
		field := g.field(in[i])
		if field.required() && field.hasFlag() && isExported(in[i].Name) {
			out[i] = in[i]
		}
	}
	return out
}

// subcommands returns the +cobra:subcommand specs declared on s, resolving each
// field's child struct from the parsed results. Specs are returned in
// deterministic field-name order. A field whose child type is not found in the
// same package is skipped.
func (g *gen) subcommands(s *parse.StructInfo) []SubcommandInfo {
	out := []SubcommandInfo{}
	fieldPrefix := g.strct(s).subcommandConfigPrefix()
	parentRequired := sortedFields(g.onlyRequired(s.Fields))
	hoisted := make([]HoistedField, 0, len(parentRequired))
	for _, f := range parentRequired {
		hoisted = append(hoisted, HoistedField{
			Field: f,
			Name:  fieldPrefix + f.Name,
		})
	}
	for k := range s.Fields {
		field := g.field(s.Fields[k])
		if !field.isSubcommand() {
			continue
		}
		child := g.lookupStruct(field.childTypeName())
		// A subcommand whose element type is not a struct in this package
		// (e.g. a []string "remove" field) has no child config; it is wired as
		// a scalar-slice subcommand with Child left nil. Only skip fields that
		// reference a struct child that genuinely could not be resolved.
		if child == nil && !field.isScalarSliceSubcommand() {
			continue
		}
		out = append(out, SubcommandInfo{
			Parent:         s,
			Field:          s.Fields[k],
			Child:          child,
			ParentRequired: parentRequired,
			TypePrefix:     s.Name,
			FieldPrefix:    fieldPrefix,
			HoistedFields:  hoisted,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Field.Name < out[j].Field.Name })
	return out
}

// lookupStruct finds a struct by name across all parsed packages.
func (g *gen) lookupStruct(name string) *parse.StructInfo {
	if g.results == nil {
		return nil
	}
	for i := range g.results.Packages {
		if s, ok := g.results.Packages[i].Structs[name]; ok {
			return s
		}
	}
	return nil
}

func onlyCobraOptions(in map[string]*parse.FieldInfo) map[string]*parse.FieldInfo {
	out := map[string]*parse.FieldInfo{}
	for i := range in {
		if _, ok := in[i].Markers["+cobra:option"]; ok {
			out[i] = in[i]
		}
	}
	return out
}

// onlyExported returns the subset of fields whose names are exported (begin
// with an uppercase letter). Unexported fields—such as the protobuf-internal
// state, sizeCache, and unknownFields fields—cannot be set from generated code
// in another package and must never appear in the generated config struct or
// To<Struct> mapping.
func onlyExported(in map[string]*parse.FieldInfo) map[string]*parse.FieldInfo {
	out := map[string]*parse.FieldInfo{}
	for i := range in {
		if isExported(in[i].Name) {
			out[i] = in[i]
		}
	}
	return out
}

func isExported(name string) bool {
	if name == "" {
		return false
	}
	r := rune(name[0])
	return r >= 'A' && r <= 'Z'
}

func (g *gen) getImports(fields map[string]*parse.FieldInfo) []string {
	imports := []string{}
	for i := range fields {
		if fields[i].IsImported && !contains(fields[i].ImportedType.ImportRaw, imports) {
			imports = append(imports, fields[i].ImportedType.ImportRaw)
		}
	}
	return imports
}

func (g *gen) getAdaptors(in *parse.Results) AdaptorInfoList {
	out := AdaptorInfoList{}
	for i := range in.Packages {
		for j := range in.Packages[i].Structs {
			for k := range in.Packages[i].Structs[j].Fields {
				field := g.field(in.Packages[i].Structs[j].Fields[k])
				if field.hasCustomConfigAdaptor() {
					inType := field.configType()
					outType := field.TypeName
					a, exists := out.getByName(field.configAdaptor())
					if exists && (a.InTypeName != inType || a.OutTypeName != outType) {
						panic(fmt.Sprintf("On field %s on struct %s.%s conflicting config adaptor definitions for %s, func(%s)(%s, error) vs func(%s)(%s, error)", field.Name, in.Packages[i].Name, in.Packages[i].Structs[j].Name, field.configAdaptor(), inType, outType, a.InTypeName, a.OutTypeName))
					}
					out = append(out, AdaptorInfo{Name: field.configAdaptor(), InTypeName: inType, OutTypeName: outType})
				}
				if field.hasCustomFlagAdaptor() {
					inType := CobraToGoType(field.flagType())
					outType := field.configType()
					current, exists := out.getByName(field.flagAdaptor())
					if exists && (current.InTypeName != inType || current.OutTypeName != outType) {
						panic(fmt.Sprintf("On field %s on struct %s.%s conflicting flag adaptor definitions for %s, func(%s)(%s, error) vs func(%s)(%s, error)", field.Name, in.Packages[i].Name, in.Packages[i].Structs[j].Name, field.flagAdaptor(), inType, outType, current.InTypeName, current.OutTypeName))
					}
					out = append(out, AdaptorInfo{Name: field.flagAdaptor(), InTypeName: inType, OutTypeName: outType})
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// sharedImports returns the imports needed by the shared adaptor-registration
// file. Adaptor in/out types may reference the source package (e.g.
// "example.MyCustomType"), so its import is included when present.
func (g *gen) sharedImports(in *parse.Results) []string {
	if g.opts.SamePackageAsSource || g.opts.SourceImport == "" {
		return nil
	}
	pkgPrefix := sourcePackageName(in)
	if pkgPrefix == "" {
		return nil
	}
	needle := pkgPrefix + "."
	for _, a := range g.getAdaptors(in) {
		if strings.Contains(a.InTypeName, needle) || strings.Contains(a.OutTypeName, needle) {
			return []string{fmt.Sprintf("%q", g.opts.SourceImport)}
		}
	}
	return nil
}

// sourcePackageName returns the name of the package that contains the
// annotated source structs. The parser may return several packages (e.g. an
// existing output package or a sibling main package) and ranges over a map, so
// we must select deterministically by preferring the package that actually
// declares structs rather than relying on map iteration order.
func sourcePackageName(in *parse.Results) string {
	best := ""
	bestStructs := -1
	for _, p := range in.Packages {
		n := len(p.Structs)
		// Prefer the package with the most structs; break ties by name to keep
		// the result stable across runs.
		if n > bestStructs || (n == bestStructs && p.Name < best) {
			best = p.Name
			bestStructs = n
		}
	}
	return best
}

func contains(value string, array []string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

// sortedFields returns the fields ordered by name for deterministic output.
func sortedFields(in map[string]*parse.FieldInfo) []*parse.FieldInfo {
	out := make([]*parse.FieldInfo, 0, len(in))
	for _, f := range in {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// requiredImports computes the full, deduplicated, sorted set of import lines
// the generated flags file needs: the standard helpers, the source package,
// the adaptors SDK, cobra, and any imported field types.
func (g *gen) requiredImports(structInfo *parse.StructInfo) []string {
	set := map[string]struct{}{
		`"fmt"`:                    {},
		`"github.com/spf13/cobra"`: {},
		`"github.com/gocloud9/gen-cobra-flags/sdk/pkg/adaptors"`: {},
	}

	// reflect is only used for DeepEqual comparisons on non-primitive fields.
	if g.needsReflect(structInfo) {
		set[`"reflect"`] = struct{}{}
	}

	// Source package import, needed for the generated To<Struct> methods.
	if !g.opts.SamePackageAsSource && g.opts.SourceImport != "" {
		set[fmt.Sprintf("%q", g.opts.SourceImport)] = struct{}{}
	}

	// Imported field types (e.g. "net", "time"). Only exported fields are
	// emitted in the generated code, so unexported fields (e.g. the
	// protobuf-internal state field of type protoimpl.MessageState) must not
	// contribute imports, otherwise the generated file fails to compile with
	// an "imported and not used" error.
	for _, f := range structInfo.Fields {
		if !isExported(f.Name) {
			continue
		}
		if f.IsImported && f.ImportedType != nil {
			set[f.ImportedType.ImportRaw] = struct{}{}
		}
	}

	out := make([]string, 0, len(set))
	for imp := range set {
		out = append(out, imp)
	}
	sort.Strings(out)
	return out
}
