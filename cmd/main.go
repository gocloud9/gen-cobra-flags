package main

import (
	"embed"
	"fmt"
	"github.com/gocloud9/gen-cobra-flags/sdk/pkg/adaptors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/gocloud9/gen-tool/pkg/generate"
	"github.com/gocloud9/gen-tool/pkg/parse"
)

//go:embed templates/*
var fs embed.FS

type Input struct {
	DestinationPackage string
}

func main() {

	parser := parse.Parser{}
	cw, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	results, err := parser.ParseDirectory(parse.Options{
		Path: filepath.Join(cw, "example/"),
		SkipFilesWithContentsRegex: []*regexp.Regexp{
			regexp.MustCompile("Generated Code with gen-cobra-flags - Do Not Edit"),
		},
	})
	if err != nil {
		panic(err)
	}

	err = generate.ExecuteWithCustom(results, generate.OptionsWithCustom[Input]{
		CustomInput: Input{
			DestinationPackage: "generated",
		},
		EmdedFS: []embed.FS{fs},
		Files: generate.Files{
			{
				TemplatePath:    "templates/flags.go.tmpl",
				DestinationPath: "example/generated/{{.Struct.Name | toSnakeCase }}_gen.go",
				Type:            generate.PerStruct,
				//FormatSource:    true,
			},
			{
				TemplatePath:    "templates/shared.go.tmpl",
				DestinationPath: "example/generated/shared.go",
				Type:            generate.Global,
			},
		},

		TemplateFuncMap: template.FuncMap{
			"toRegisterMethod": func(f *AdaptorInfo) string {
				return fmt.Sprintf("Register%s%s", strings.ToUpper(f.Name)[:1], f.Name[1:])
			},
			"fieldToFlagMethod": func(f *parse.FieldInfo) string {
				field := (*Field)(f)

				if field.TypeName == "time.Time" {
					return fmt.Sprintf("%sP(\"%s\", \"%s\", %s, []string{time.RFC3339}, \"%s\")", field.flagType(), field.flag(), field.Short(), field.Default(), field.Usage())
				}

				return fmt.Sprintf("%sP(\"%s\", \"%s\", %s, \"%s\")", field.flagType(), field.flag(), field.Short(), field.Default(), field.Usage())

			},
			"fieldToFlagGetMethod": func(f *parse.FieldInfo) string {
				field := (*Field)(f)

				return fmt.Sprintf("Get%s(\"%s\")", GoToCobraType(field.TypeName), field.flag())
			},
			"getCobraTags": func(field *parse.FieldInfo) string {
				tags := []string{}
				cobraJsonTags, _ := field.Markers["+cobra:json"]
				if cobraJsonTags != "" {
					tags = append(tags, fmt.Sprintf("json:\"%s\"", cobraJsonTags))
				}

				cobraYamlTags, _ := field.Markers["+cobra:yaml"]
				if cobraYamlTags != "" {
					tags = append(tags, fmt.Sprintf("yaml:\"%s\"", cobraYamlTags))
				}

				cobraCustomTags, _ := field.Markers["+cobra:customTags"]
				if cobraCustomTags != "" {
					tags = append(tags, cobraCustomTags)
				}

				if len(tags) == 0 {
					return ""
				}

				return fmt.Sprintf(" `%s`", strings.Join(tags, " "))
			},
			"onlyCobraFlags": func(in map[string]*parse.FieldInfo) map[string]*parse.FieldInfo {
				out := map[string]*parse.FieldInfo{}
				for i := range in {
					field := (*Field)(in[i])
					if field.hasFlag() {
						out[i] = in[i]
					}
				}

				return out
			},

			"getAdaptors": func(in *parse.Results) AdaptorInfoList {
				out := AdaptorInfoList{}
				for i := range in.Packages {
					for j := range in.Packages[i].Structs {
						for k := range in.Packages[i].Structs[j].Fields {
							field := (*Field)(in.Packages[i].Structs[j].Fields[k])
							if field.hasCustomConfigAdaptor() {
								inType := field.configType()
								outType := field.TypeName

								a, exists := out.getByName(field.configAdaptor())
								if exists {
									if a.InTypeName != inType || a.OutTypeName != outType {
										panic(fmt.Sprintf("On field %s on struct %s.%s conflicting config adaptor definitions for %s, func(%s)(%s, error) vs func(%s)(%s, error)", field.Name, in.Packages[i].Name, in.Packages[i].Structs[j].Name, field.configAdaptor(), inType, outType, a.InTypeName, a.OutTypeName))
									}
								}

								out = append(out, AdaptorInfo{
									Name:        field.configAdaptor(),
									InTypeName:  inType,
									OutTypeName: outType,
								})
							}

							if field.hasCustomFlagAdaptor() {
								inType := CobraToGoType(field.flagType())
								outType := field.configType()

								current, exists := out.getByName(field.flagAdaptor())
								if exists {
									if current.InTypeName != inType || current.OutTypeName != outType {
										panic(fmt.Sprintf("On field %s on struct %s.%s conflicting flag adaptor definitions for %s, func(%s)(%s, error) vs func(%s)(%s, error)", field.Name, in.Packages[i].Name, in.Packages[i].Structs[j].Name, field.flagAdaptor(), inType, outType, current.InTypeName, current.OutTypeName))
									}
								}

								out = append(out, AdaptorInfo{
									Name:        field.flagAdaptor(),
									InTypeName:  inType,
									OutTypeName: outType,
								})
							}
						}
					}
				}

				return out
			},

			"onlyCobraOptions": func(in map[string]*parse.FieldInfo) map[string]*parse.FieldInfo {
				out := map[string]*parse.FieldInfo{}
				for i := range in {
					_, ok := in[i].Markers["+cobra:option"]
					if ok {
						out[i] = in[i]
					}
				}

				return out
			},

			"getImports": func(fields map[string]*parse.FieldInfo) []string {
				imports := []string{}
				for i := range fields {
					if fields[i].IsImported && !contains(fields[i].ImportedType.ImportRaw, imports) {
						imports = append(imports, fields[i].ImportedType.ImportRaw)
					}
				}

				return imports
			},

			"asConfigAdaptorName": func(info *parse.FieldInfo) string {
				field := (*Field)(info)

				return field.configAdaptor()
			},

			"asFlagAdaptorName": func(info *parse.FieldInfo) string {
				field := (*Field)(info)

				return field.flagAdaptor()
			},

			"fieldToConfigTypeName": func(field *parse.FieldInfo) string {
				f := (*Field)(field)

				return f.configType()
			},

			"structToFlagAdaptorName": func(structInfo *parse.StructInfo) string {
				s := (*Struct)(structInfo)

				return s.flagAdaptor()
			},

			"structToFlagMethod": func(structInfo *parse.StructInfo) string {
				s := (*Struct)(structInfo)

				return fmt.Sprintf("StringP(\"%s\", \"%s\", %s, \"%s\")", s.flag(), s.Short(), s.Default(), s.Usage())
			},

			"structToFlagGetMethod": func(structInfo *parse.StructInfo) string {
				s := (*Struct)(structInfo)

				return fmt.Sprintf("GetString(\"%s\")", s.flag())
			},

			"needsConfigAdaptor": func(f *parse.FieldInfo) bool {
				field := (*Field)(f)
	
				_, ok := field.Markers["+cobra:config:adaptor"]
				if ok {
					return true
				}

				return false
			},

			"toSnakeCase": func(in string) string {
				return strings.ToLower(regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(in, "${1}_${2}"))
			},
		},
	})

	if err != nil {
		panic(err)
	}

}

type AdaptorInfoList []AdaptorInfo

func (a AdaptorInfoList) getByName(adaptorName string) (AdaptorInfo, bool) {
	for i := range a {
		if a[i].Name == adaptorName {
			return a[i], true
		}
	}

	return AdaptorInfo{}, false

}

type AdaptorInfo struct {
	Name        string
	InTypeName  string
	OutTypeName string
}

func contains(value string, array []string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

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
		foundType, found := strings.CutPrefix(typeName, "[]")
		if found {
			return GoToCobraType(foundType) + "Slice"
		}
		foundType, found = strings.CutPrefix(typeName, "map[string]")
		if found {
			return "StringTo" + GoToCobraType(foundType)
		}

		if strings.ToUpper(typeName[:1]) == typeName[:1] {
			return "String"
		}

		return strings.ToUpper(typeName[:1]) + typeName[1:]
	}
}

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
		foundType, found := strings.CutSuffix(typeName, "Slice")
		if found {
			return "[]" + CobraToGoType(foundType)
		}

		foundType, found = strings.CutPrefix(typeName, "StringTo")
		if found {
			return "map[string]" + CobraToGoType(foundType)
		}

		return strings.ToLower(typeName)
	}
}

type Struct parse.StructInfo

func (s *Struct) hasFlag() bool {
	_, ok := s.Markers["+cobra:flag"]

	return ok
}

func (s *Struct) flagAdaptor() string {
	a, ok := s.Markers["+cobra:flag:adaptor"]
	if ok {
		return a
	}

	return "adaptors.JsonOrYamlToStruct[" + s.Name + "Config]"
}

func (s *Struct) flag() string {
	cobraFlag, _ := s.Markers["+cobra:flag"]

	return cobraFlag
}

func (s *Struct) Short() string {
	cobraShort, _ := s.Markers["+cobra:short"]
	return cobraShort
}

func (s *Struct) Usage() string {
	cobraUsage, _ := s.Markers["+cobra:usage"]
	return cobraUsage
}

func (s *Struct) Default() string {
	cobraDefault, ok := s.Markers["+cobra:default"]
	if !ok {
		return "\"\""
	}

	return cobraDefault
}

type Field parse.FieldInfo

func (f *Field) configAdaptor() string {
	a, ok := f.Markers["+cobra:config:adaptor"]
	if !ok {
		inType := f.configType()
		outType := f.TypeName

		return adaptors.GetFuncNameByTypeNames(inType, outType)
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

		return adaptors.GetFuncNameByTypeNames(inType, outType)
	}

	return "adaptor" + a
}

func (f *Field) hasCustomFlagAdaptor() bool {
	_, ok := f.Markers["+cobra:flag:adaptor"]
	return ok
}

func (f *Field) flagType() string {
	t, ok := f.Markers["+cobra:flag:type"]
	if !ok {
		return GoToCobraType(f.configType())
	}

	return t
}

func (f *Field) configType() string {
	t, ok := f.Markers["+cobra:config:type"]
	if !ok {
		return f.TypeName
	}

	return t
}

func (f *Field) flag() string {
	cobraFlag, _ := f.Markers["+cobra:flag"]

	return cobraFlag
}

func (f *Field) hasFlag() bool {
	_, ok := f.Markers["+cobra:flag"]

	return ok
}

func (f *Field) Short() string {
	cobraShort, _ := f.Markers["+cobra:short"]
	return cobraShort
}

func (f *Field) Usage() string {
	cobraUsage, _ := f.Markers["+cobra:usage"]
	return cobraUsage
}

func (f *Field) Default() string {
	cobraDefault, _ := f.Markers["+cobra:default"]
	return cobraDefault
}
