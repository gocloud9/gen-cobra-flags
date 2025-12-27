package main

import (
	"embed"
	"fmt"
	"github.com/gocloud9/gen-tool/pkg/generate"
	"github.com/gocloud9/gen-tool/pkg/parse"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*
var fs embed.FS

func main() {

	parser := parse.Parser{}
	cw, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	results, err := parser.ParseDirectory(filepath.Join(cw, "example/"))
	if err != nil {
		panic(err)
	}

	err = generate.Execute(results, generate.Options{
		EmdedFS: []embed.FS{fs},
		Files: generate.Files{
			{
				TemplatePath:    "templates/flags.go.tmpl",
				DestinationPath: "{{.Struct.Name}}_gen.go",
				Type:            generate.PerStruct,
				//FormatSource:    true,
			},
			{
				TemplatePath:    "templates/shared.go.tmpl",
				DestinationPath: "{{.Package.Name}}_gen_shared.go",
				Type:            generate.PerPackage,
			},
		},
		TemplateFuncMap: template.FuncMap{
			"asCobraFlag": func(field *parse.FieldInfo) string {
				typeName := field.TypeName
				cobraType, ok := field.Markers["+cobra:type"]
				if !ok {
					cobraType = GoToCobraType(typeName)
				}

				cobraFlag, ok := field.Markers["+cobra:flag"]
				if !ok {
					panic("cobra:flag is missing")
				}

				cobraShort, _ := field.Markers["+cobra:short"]
				cobraUsage, _ := field.Markers["+cobra:usage"]
				cobraDefault, _ := field.Markers["+cobra:default"]

				return fmt.Sprintf("cmd.Flags().%sP(\"%s\", \"%s\", %s, \"%s\")", cobraType, cobraFlag, cobraShort, cobraDefault, cobraUsage)

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
					_, ok := in[i].Markers["+cobra:flag"]
					if ok {
						out[i] = in[i]
					}
				}

				return out
			},

			"getCobraAdaptors": func(in *parse.PackageInfo) []AdaptorInfo {
				out := []AdaptorInfo{}
				for j := range in.Structs {
					for k := range in.Structs[j].Fields {
						aName, ok := in.Structs[j].Fields[k].Markers["+cobra:adaptor"]
						if ok {
							cobraType, ok := in.Structs[j].Fields[k].Markers["+cobra:type"]
							if !ok {
								cobraType = GoToCobraType(in.Structs[j].Fields[k].TypeName)
							}

							out = append(out, AdaptorInfo{
								Name:        aName,
								InTypeName:  cobraType,
								OutTypeName: in.Structs[j].Fields[k].TypeName,
							})
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
		},
	})

	if err != nil {
		panic(err)
	}

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
	case "net.IPNet":
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
			return "string"
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
