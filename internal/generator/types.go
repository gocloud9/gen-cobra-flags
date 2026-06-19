package generator

import (
	"strings"

	"github.com/gocloud9/gen-tool/pkg/parse"

	"github.com/gocloud9/gen-cobra-flags/sdk/pkg/adaptors"
)

// AdaptorInfo describes a generated adaptor registration hook.
type AdaptorInfo struct {
	Name        string
	InTypeName  string
	OutTypeName string
}

// AdaptorInfoList is a collection of AdaptorInfo with lookup helpers.
type AdaptorInfoList []AdaptorInfo

func (a AdaptorInfoList) getByName(adaptorName string) (AdaptorInfo, bool) {
	for i := range a {
		if a[i].Name == adaptorName {
			return a[i], true
		}
	}
	return AdaptorInfo{}, false
}

// SubcommandInfo is the template view-model for a +cobra:subcommand field. It
// captures everything the parent's generated file needs to wire a subcommand
// for the singular form of a repeated resource: the parent struct, the field,
// the resolved child struct, and the parent's required fields (which are also
// surfaced as flags on the subcommand).
//
// The generated output is a composed config struct that embeds the parent's
// required ("hoisted") fields alongside the child config, e.g. for parent
// UpdateVpcRequest and child CreateVpcPeer:
//
//	type UpdateVpcRequestChildRequiredFields struct {
//	    VpcName string
//	}
//	type UpdateVpcRequestCreateVpcPeerConfig struct {
//	    UpdateVpcRequestChildRequiredFields `yaml:",inline"`
//	    CreateVpcPeerConfig                 `json:",inline"`
//	}
//	func UpdateVpcRequestCreateVpcPeerConfigFromFlags(cmd) (*UpdateVpcRequestCreateVpcPeerConfig, error)
type SubcommandInfo struct {
	Parent         *parse.StructInfo
	Field          *parse.FieldInfo
	Child          *parse.StructInfo
	ParentRequired []*parse.FieldInfo

	// TypePrefix is the prefix applied to generated type names. It is the
	// parent message name verbatim (e.g. "UpdateVpcRequest"), so the composed
	// config is "<TypePrefix><Child>Config" and the hoisted-fields struct is
	// "<TypePrefix>ChildRequiredFields".
	TypePrefix string

	// FieldPrefix is the prefix applied to hoisted parent field names, taken
	// from the parent's +cobra:subcommand:config:prefix marker (e.g. "Vpc").
	// A hoisted field is named FieldPrefix + parentGoFieldName, so parent
	// field "Name" becomes "VpcName".
	FieldPrefix string

	// HoistedFields are the parent's required fields surfaced in the composed
	// config's ChildRequiredFields struct, carrying their renamed identifiers.
	HoistedFields []HoistedField
}

// configStructName returns the name of the composed config struct for the
// subcommand, e.g. "UpdateVpcRequestCreateVpcPeerConfig". For a scalar-slice
// subcommand (no struct child, e.g. a []string field) it is named from the
// singular field instead, e.g. "UpdateVpcRequestRemovePeerConfig".
func (s SubcommandInfo) configStructName() string {
	if s.isScalarSlice() {
		return s.TypePrefix + s.scalarFieldName() + "Config"
	}
	return s.TypePrefix + s.Child.Name + "Config"
}

// isScalarSlice reports whether the subcommand operates on a repeated field of
// a primitive element type (e.g. []string) rather than a struct child. Such a
// subcommand has no <Child>Config / To<Child> machinery: it surfaces a single
// scalar value flag whose value is appended to the parent's slice field. It is
// detected by the absence of a resolved Child struct.
func (s SubcommandInfo) isScalarSlice() bool {
	return s.Child == nil
}

// scalarFieldName returns the singular Go field name used in the composed
// config for a scalar-slice subcommand, e.g. field "RemovePeers" -> "RemovePeer".
func (s SubcommandInfo) scalarFieldName() string {
	return subcommandName(s.Field.Name)
}

// scalarValueFlag returns the flag name for the single scalar value of a
// scalar-slice subcommand, taken from +cobra:subcommand:value:flag, falling
// back to a kebab-cased singular field name (e.g. "remove-peer").
func (s SubcommandInfo) scalarValueFlag() string {
	if v, ok := s.Field.Markers["+cobra:subcommand:value:flag"]; ok && v != "" {
		return v
	}
	return toKebabCase(s.scalarFieldName())
}

// scalarValueShort returns the short flag for the scalar value of a
// scalar-slice subcommand, from +cobra:subcommand:value:short (empty when absent).
func (s SubcommandInfo) scalarValueShort() string {
	return s.Field.Markers["+cobra:subcommand:value:short"]
}

// scalarValueUsage returns the usage string for the scalar value of a
// scalar-slice subcommand, from +cobra:subcommand:value:usage, falling back to
// a generic description. Required-ness is enforced post-merge, so the usage is
// prefixed with "[required] ".
func (s SubcommandInfo) scalarValueUsage() string {
	if v, ok := s.Field.Markers["+cobra:subcommand:value:usage"]; ok && v != "" {
		return "[required] " + v
	}
	return "[required] The " + s.scalarFieldName() + " value."
}

// requiredFieldsStructName returns the name of the per-parent struct that holds
// the hoisted required fields, e.g. "UpdateVpcRequestChildRequiredFields".
func (s SubcommandInfo) requiredFieldsStructName() string {
	return s.TypePrefix + "ChildRequiredFields"
}

// childConfigName returns the child's own config struct name, e.g.
// "CreateVpcPeerConfig".
func (s SubcommandInfo) childConfigName() string {
	return s.Child.Name + "Config"
}

// aggregateFlag returns the name of the composed-config aggregate flag for the
// subcommand, taken from the parent field's +cobra:subcommand:config:flag
// marker. When absent it falls back to a kebab-cased "<child>-config" name.
func (s SubcommandInfo) aggregateFlag() string {
	if v, ok := s.Field.Markers["+cobra:subcommand:config:flag"]; ok && v != "" {
		return v
	}
	return toKebabCase(s.Child.Name) + "-config"
}

// aggregateShort returns the short flag for the composed-config aggregate flag,
// taken from the +cobra:subcommand:config:short marker (empty when absent).
func (s SubcommandInfo) aggregateShort() string {
	return s.Field.Markers["+cobra:subcommand:config:short"]
}

// aggregateUsage returns the usage string for the composed-config aggregate
// flag, taken from the +cobra:subcommand:config:usage marker. When absent it
// falls back to a generic description.
func (s SubcommandInfo) aggregateUsage() string {
	if v, ok := s.Field.Markers["+cobra:subcommand:config:usage"]; ok && v != "" {
		return v
	}
	return "Aggregate JSON/YAML config for " + s.configStructName() + "."
}

// parentTypeRef returns the parent message type as referenced from the
// generated package: bare when generating into the source package, otherwise
// qualified with the source package name (e.g. "types.UpdateVpcRequest").
func (s SubcommandInfo) parentTypeRef(samePackage bool, pkgName string) string {
	if samePackage {
		return s.Parent.Name
	}
	return pkgName + "." + s.Parent.Name
}

// childToMethod returns the child config's conversion method name that yields
// the child domain type, e.g. "ToCreateVpcPeer".
func (s SubcommandInfo) childToMethod() string {
	return "To" + s.Child.Name
}

// singularFieldName returns the parent field that the subcommand operates on,
// e.g. "AddPeers". The final conversion writes the single converted child into
// this field on the parent domain struct.
func (s SubcommandInfo) singularFieldName() string {
	return s.Field.Name
}

// singularFieldIsSlice reports whether the parent field is a repeated resource,
// so the final conversion wraps the single converted child in a one-element
// slice (vs assigning a scalar pointer directly).
func (s SubcommandInfo) singularFieldIsSlice() bool {
	return s.Field.IsSlice
}

// HoistedField is a parent required field surfaced in the composed config's
// ChildRequiredFields struct under a prefixed name.
type HoistedField struct {
	// Field is the original parent field (used for flag get/type helpers).
	Field *parse.FieldInfo
	// Name is the renamed identifier in the hoisted struct (FieldPrefix +
	// original field name, e.g. "VpcName").
	Name string
}

// GoToCobraType maps a Go type name to the corresponding pflag/cobra accessor
// suffix (e.g. "time.Duration" -> "Duration", "[]string" -> "StringSlice").
func GoToCobraType(typeName string) string {
	switch typeName {
	case "time.Duration":
		return "Duration"
	case "time.Time":
		return "Time"
	case "*net.IPNet":
		return "IPNet"
	case "net.IP":
		return "IP"
	default:
		if foundType, found := strings.CutPrefix(typeName, "[]"); found {
			return GoToCobraType(foundType) + "Slice"
		}
		if foundType, found := strings.CutPrefix(typeName, "map[string]"); found {
			return "StringTo" + GoToCobraType(foundType)
		}
		if strings.ToUpper(typeName[:1]) == typeName[:1] {
			return "String"
		}
		if strings.Contains(typeName, ".") {
			return "String"
		}
		return strings.ToUpper(typeName[:1]) + typeName[1:]
	}
}

// CobraToGoType is the inverse of GoToCobraType.
func CobraToGoType(typeName string) string {
	switch typeName {
	case "Duration":
		return "time.Duration"
	case "Time":
		return "time.Time"
	case "IPNet":
		return "net.IPNet"
	case "IP":
		return "net.IP"
	default:
		if foundType, found := strings.CutSuffix(typeName, "Slice"); found {
			return "[]" + CobraToGoType(foundType)
		}
		if foundType, found := strings.CutSuffix(typeName, "Array"); found {
			return "[]" + CobraToGoType(foundType)
		}
		if foundType, found := strings.CutPrefix(typeName, "StringTo"); found {
			return "map[string]" + CobraToGoType(foundType)
		}
		return strings.ToLower(typeName)
	}
}

// GetBaseType walks a type's underlying chain until it reaches a struct or a
// non-defined type.
func GetBaseType(t *parse.TypeInfo) *parse.TypeInfo {
	if t.IsStruct {
		return t
	}
	if t.IsType {
		return GetBaseType(t.TypeOf)
	}
	return t
}

// Struct wraps parse.StructInfo with marker-aware helpers.
type Struct struct {
	*parse.StructInfo
}

func (s *Struct) flagAdaptor() string {
	if a, ok := s.Markers["+cobra:flag:adaptor"]; ok {
		return a
	}
	return "adaptors.JsonOrYamlToStruct[" + s.Name + "Config]"
}

func (s *Struct) flag() string  { return s.Markers["+cobra:flag"] }
func (s *Struct) short() string { return s.Markers["+cobra:short"] }
func (s *Struct) usage() string { return s.Markers["+cobra:usage"] }

// enabled reports whether the struct carries a +cobra:enabled marker. Selection
// for generation is driven by this marker (replacing the former -struct CLI
// flag): every +cobra:enabled struct is generated, along with any child structs
// pulled in by +cobra:config:child fields. An explicit "false" disables it.
func (s *Struct) enabled() bool {
	v, ok := s.Markers["+cobra:enabled"]
	if !ok {
		return false
	}
	return v != "false"
}

// hasFlag reports whether the struct declares an aggregate +cobra:flag. Child
// structs synthesized for +cobra:config:child have no aggregate flag, so the
// generated config-from-flags path must skip reading/decoding it.
func (s *Struct) hasFlag() bool {
	_, ok := s.Markers["+cobra:flag"]
	return ok
}

// subcommandConfigPrefix returns the prefix used to rename hoisted parent
// fields in a subcommand's composed config struct, taken from the parent's
// +cobra:subcommand:config:prefix marker. For example a prefix of "Vpc" renames
// the parent's "Name" field to "VpcName". An empty result means no prefix.
func (s *Struct) subcommandConfigPrefix() string {
	return s.Markers["+cobra:subcommand:config:prefix"]
}

func (s *Struct) defaultValue() string {
	if v, ok := s.Markers["+cobra:default"]; ok {
		return v
	}
	return `""`
}

// Field wraps parse.FieldInfo with marker-aware helpers. It carries the
// per-run samePackageAsSource flag so type references resolve correctly.
type Field struct {
	*parse.FieldInfo
	samePackageAsSource bool
}

func (f *Field) configAdaptor() string {
	a, ok := f.Markers["+cobra:config:adaptor"]
	if !ok {
		return ""
	}
	return "adaptor" + a
}

func (f *Field) hasCustomConfigAdaptor() bool {
	_, ok := f.Markers["+cobra:config:adaptor"]
	return ok
}

func (f *Field) flagAdaptor() string {
	a, ok := f.Markers["+cobra:flag:adaptor"]
	if !ok {
		inType := CobraToGoType(f.flagType())
		outType := f.configType()
		if "*"+inType == outType {
			return "adaptors.ToPtr"
		}
		// Complex (non-primitive) fields have no direct CLI representation, so
		// the generated flag is a string (scalar) or string slice (collection)
		// of JSON/YAML values decoded per element into the config type.
		if f.isStructLikeSlice() {
			return "adaptors.StringSliceToStructSlice[" + f.sliceElemConfigType() + "]"
		}
		if f.isStructLikeScalar() {
			return "adaptors.StringToStruct[" + outType + "]"
		}
		name := adaptors.GetFuncNameByTypeNames(inType, outType)
		if name == "" {
			return "adaptors.ToPtr"
		}
		return "adaptors." + name
	}
	return "adaptor" + a
}

// isStructLikeScalar reports whether the field is a struct or pointer-to-struct
// config value that must be hydrated from a single JSON/YAML string flag.
func (f *Field) isStructLikeScalar() bool {
	if f.IsSlice || f.IsMap {
		return false
	}
	return underlyingIsStruct(f.TypeInfo)
}

// isStructLikeSlice reports whether the field is a slice whose element is a
// struct or pointer-to-struct, so each element is decoded from a JSON/YAML
// string in a StringSlice flag.
func (f *Field) isStructLikeSlice() bool {
	if !f.IsSlice || f.Slice == nil {
		return false
	}
	return underlyingIsStruct(f.Slice)
}

// underlyingIsStruct reports whether t, after unwrapping pointers and named
// type definitions, is a struct type. Such types have no primitive CLI
// representation and must be decoded from JSON/YAML.
func underlyingIsStruct(t *parse.TypeInfo) bool {
	for t != nil {
		switch {
		case t.IsStruct:
			return true
		case t.IsPointer:
			t = t.Pointer
		case t.IsType:
			t = t.TypeOf
		default:
			return false
		}
	}
	return false
}

// sliceElemConfigType returns the element type of a slice config field,
// respecting the same/external package qualification used by configType. For
// example "[]*CreateVpcSubnet" yields "*CreateVpcSubnet".
func (f *Field) sliceElemConfigType() string {
	if elem, ok := strings.CutPrefix(f.configType(), "[]"); ok {
		return elem
	}
	return f.configType()
}

func (f *Field) hasCustomFlagAdaptor() bool {
	_, ok := f.Markers["+cobra:flag:adaptor"]
	return ok
}

func (f *Field) flagType() string {
	if t, ok := f.Markers["+cobra:flag:type"]; ok {
		return t
	}
	return GoToCobraType(f.configType())
}

func (f *Field) configType() string {
	if t, ok := f.Markers["+cobra:config:type"]; ok {
		return t
	}
	// A +cobra:config:child field is represented in the parent config by the
	// generated child Config type (e.g. "[]*CreateVpcPeerConfig"), not the
	// source domain type. The child config lives in the generated package, so
	// any source-package qualifier is dropped.
	if f.isConfigChild() {
		return f.configChildType()
	}
	if f.samePackageAsSource {
		return f.TypeName
	}
	return f.ExternalTypeName
}

// configChildType returns the parent-config representation of a
// +cobra:config:child field: the field's type with its element struct replaced
// by the generated "<Child>Config" type and any package qualifier stripped. The
// slice/pointer decoration is preserved, so "[]*pkg.CreateVpcPeer" becomes
// "[]*CreateVpcPeerConfig" and "*pkg.CreateVpcPeer" becomes "*CreateVpcPeerConfig".
func (f *Field) configChildType() string {
	configElem := f.childTypeName() + "Config"
	prefix := ""
	if f.IsSlice {
		prefix += "[]"
	}
	if f.IsPointer || (f.IsSlice && f.Slice != nil && f.Slice.IsPointer) {
		prefix += "*"
	}
	return prefix + configElem
}

func (f *Field) configIsSameType() bool {
	t, ok := f.Markers["+cobra:config:type"]
	if !ok {
		return true
	}
	return t == f.TypeName
}

func (f *Field) flag() string  { return f.Markers["+cobra:flag"] }
func (f *Field) short() string { return f.Markers["+cobra:short"] }
func (f *Field) usage() string { return f.Markers["+cobra:usage"] }

// isConfigChild reports whether the field carries a +cobra:config:child marker,
// which requests that the field's element struct be generated as its own child
// config (config struct + flags + To<Child>). An explicit "false" disables it.
func (f *Field) isConfigChild() bool {
	v, ok := f.Markers["+cobra:config:child"]
	if !ok {
		return false
	}
	return v != "false"
}

// isSubcommand reports whether the field carries a +cobra:subcommand marker,
// which requests subcommand wiring for the singular form of a repeated resource
// field (child flags plus the parent's required flags). An explicit "false"
// disables it.
func (f *Field) isSubcommand() bool {
	v, ok := f.Markers["+cobra:subcommand"]
	if !ok {
		return false
	}
	return v != "false"
}

// isScalarSliceSubcommand reports whether a +cobra:subcommand field is a
// repeated field of a primitive element type (e.g. []string) rather than a
// struct child. Such a subcommand surfaces a single scalar value flag whose
// value is appended to the parent's slice field, with no <Child>Config or
// To<Child> conversion.
func (f *Field) isScalarSliceSubcommand() bool {
	if !f.IsSlice {
		return false
	}
	if f.Slice != nil && underlyingIsStruct(f.Slice) {
		return false
	}
	return isPrimitive(f.childTypeName())
}

// childTypeName returns the element struct name referenced by a config:child or
// subcommand field, with any slice/pointer decoration stripped. For example
// "[]*CreateVpcPeer" yields "CreateVpcPeer".
func (f *Field) childTypeName() string {
	name := f.TypeName
	if f.IsSlice && f.Slice != nil {
		name = f.Slice.TypeName
	}
	name = strings.TrimPrefix(name, "[]")
	name = strings.TrimPrefix(name, "*")
	return name
}

// domainType returns the source (domain) Go type of the field, including any
// package qualifier — i.e. the type produced by To<Struct> regardless of any
// config-child substitution applied to the config struct field. For example
// "[]*types.CreateVpcPeer". It mirrors the pre-config:child behavior of
// configType.
func (f *Field) domainType() string {
	if f.samePackageAsSource {
		return f.TypeName
	}
	return f.ExternalTypeName
}

// toChildMethod returns the name of the generated conversion method on the
// child config that produces the domain type, e.g. "ToCreateVpcPeer".
func (f *Field) toChildMethod() string {
	return "To" + f.childTypeName()
}

// configChildIsSlice reports whether a config:child field is a slice (vs a
// scalar struct/pointer), which determines whether To<Struct> emits an
// element-wise conversion loop or a single conversion.
func (f *Field) configChildIsSlice() bool {
	return f.IsSlice
}

// required reports whether the field carries a +cobra:required marker. The
// marker value is ignored except that an explicit "false" disables it, so both
// "+cobra:required" and "+cobra:required=true" mark the field as required.
func (f *Field) required() bool {
	v, ok := f.Markers["+cobra:required"]
	if !ok {
		return false
	}
	return v != "false"
}

// flagUsage returns the usage string registered with the flag. Required fields
// are prefixed with "[required] " because requiredness is enforced after
// merging the aggregate config flag and individual flags (not via cobra's own
// MarkFlagRequired), so the help text must signal it explicitly.
func (f *Field) flagUsage() string {
	if f.required() {
		return "[required] " + f.usage()
	}
	return f.usage()
}

func (f *Field) hasFlag() bool {
	_, ok := f.Markers["+cobra:flag"]
	return ok
}

func (f *Field) flagDefault() string {
	if v, ok := f.Markers["+cobra:default"]; ok {
		return v
	}
	return f.flagZero()
}

func (f *Field) configZero() string {
	if z, ok := primitiveZero(f.configType()); ok {
		return z
	}
	if f.IsPointer || f.IsMap || f.IsSlice {
		return "nil"
	}
	if f.IsStruct {
		return f.configType() + "{}"
	}

	baseType := GetBaseType(f.TypeInfo)
	if z, ok := primitiveZero(baseType.TypeName); ok {
		return z
	}
	if f.IsPointer || f.IsMap || f.IsSlice {
		return "nil"
	}
	if f.IsStruct {
		return f.configType() + "{}"
	}
	return "nil"
}

func (f *Field) flagZero() string {
	switch f.flagType() {
	case "String":
		return `""`
	case "Bool":
		return "false"
	case "Int", "Int8", "Int16", "Int32", "Int64",
		"Uint", "Uint8", "Uint16", "Uint32", "Uint64",
		"Float32", "Float64":
		return "0"
	case "Duration":
		return "0"
	case "Time":
		return "time.Time{}"
	case "IP":
		return "net.IP{}"
	case "IPNet":
		return f.TypeName + "{}"
	default:
		return "nil"
	}
}

func primitiveZero(typeName string) (string, bool) {
	switch typeName {
	case "string":
		return `""`, true
	case "bool":
		return "false", true
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		return "0", true
	default:
		return "", false
	}
}
