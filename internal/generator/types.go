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
		name := adaptors.GetFuncNameByTypeNames(inType, outType)
		if name == "" {
			return "adaptors.ToPtr"
		}
		return "adaptors." + name
	}
	return "adaptor" + a
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
	if f.samePackageAsSource {
		return f.TypeName
	}
	return f.ExternalTypeName
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
