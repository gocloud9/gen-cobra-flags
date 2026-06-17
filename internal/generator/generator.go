// Package generator implements the gen-cobra-flags code generator.
//
// It parses Go source annotated with +cobra:* markers and produces Cobra
// flag-binding boilerplate. The public entry point is Generate.
package generator

import (
	"embed"
	"fmt"
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

	// When a specific struct is requested, filter the parse results down to it.
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

// gen holds per-run state and the template helper methods.
type gen struct {
	opts Options
}

func (g *gen) funcMap() template.FuncMap {
	return template.FuncMap{
		"toRegisterMethod": func(f *AdaptorInfo) string {
			return fmt.Sprintf("Register%s%s", strings.ToUpper(f.Name)[:1], f.Name[1:])
		},
		"fieldToFlagMethod": func(f *parse.FieldInfo) string {
			field := g.field(f)
			if field.TypeName == "time.Time" {
				return fmt.Sprintf("%sP(%q, %q, %s, []string{time.RFC3339}, %q)", field.flagType(), field.flag(), field.short(), field.flagDefault(), field.usage())
			}
			return fmt.Sprintf("%sP(%q, %q, %s, %q)", field.flagType(), field.flag(), field.short(), field.flagDefault(), field.usage())
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

func toSnakeCase(in string) string {
	return strings.ToLower(regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(in, "${1}_${2}"))
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

func onlyCobraOptions(in map[string]*parse.FieldInfo) map[string]*parse.FieldInfo {
	out := map[string]*parse.FieldInfo{}
	for i := range in {
		if _, ok := in[i].Markers["+cobra:option"]; ok {
			out[i] = in[i]
		}
	}
	return out
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

	// Imported field types (e.g. "net", "time").
	for _, f := range structInfo.Fields {
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
