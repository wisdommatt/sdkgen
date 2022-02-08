package openapi

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/imports"
	"gopkg.in/yaml.v2"
)

type Property struct {
	Description          string              `json:"description" yaml:"description"`
	Properties           map[string]Property `json:"properties" yaml:"properties"`
	Required             []string            `json:"required" yaml:"required"`
	Type                 string              `json:"type" yaml:"type"`
	XGoPackage           string              `json:"x-go-package" yaml:"x-go-package"`
	Ref                  string              `json:"$ref" yaml:"$ref"`
	Format               string              `json:"format" yaml:"format"`
	XGoName              string              `json:"x-go-name" yaml:"x-go-name"`
	AdditionalProperties *Property           `json:"additionalProperties" yaml:"additionalProperties"`
	Items                *Property           `json:"items" yaml:"items"`
	XML                  struct {
		Name    string `json:"name"`
		Wrapped bool   `json:"wrapped"`
	} `json:"xml"`
	Default struct {
		Description string `json:"description"`
	} `json:"default"`
	Schema *Property `json:"schema" yaml:"schema"`
}

type Info struct {
	Description string `json:"description" yaml:"description"`
	Title       string `json:"title" yaml:"title"`
	Version     string `json:"version" yaml:"version"`
}

type Path struct {
	Description string                `json:"description" yaml:"description"`
	OperationID string                `json:"operationId" yaml:"operationId"`
	Parameters  []PathParameter       `json:"parameters" yaml:"parameters"`
	Responses   map[string]Property   `json:"responses" yaml:"responses"`
	Summary     string                `json:"summary" yaml:"summary"`
	Tags        []string              `json:"tags" yaml:"tags"`
	Security    []map[string][]string `json:"security" yaml:"security"`
	Schemes     []string              `json:"schemes" yaml:"schemes"`
	Consumes    []string              `json:"consumes" yaml:"consumes"`
	Produces    []string              `json:"produces"`
}

type PathParameter struct {
	Description string   `json:"description" yaml:"description"`
	In          string   `json:"in" yaml:"in"`
	Name        string   `json:"name" yaml:"name"`
	Required    bool     `json:"required" yaml:"required"`
	Schema      Property `json:"schema" yaml:"schema"`
}
type Response struct {
	Description string   `json:"description" yaml:"description"`
	Schema      Property `json:"schema" yaml:"schema"`
}

type OpenAPISchema struct {
	BasePath            string                        `json:"basePath" yaml:"basePath"`
	Consumes            []string                      `json:"consumes" yaml:"consumes"`
	Definitions         map[string]Property           `json:"definitions" yaml:"definitions"`
	Host                string                        `json:"host" yaml:"host"`
	Info                Info                          `json:"info" yaml:"info"`
	Paths               map[string]map[string]Path    `json:"paths" yaml:"paths"`
	Produces            []string                      `json:"produces" yaml:"produces"`
	Responses           map[string]Property           `json:"responses" yaml:"responses"`
	Schemes             []string                      `json:"schemes" yaml:"schemes"`
	SecurityDefinitions map[string]SecurityDefinition `json:"securityDefinitions" yaml:"securityDefinitions"`
	Swagger             string                        `json:"swagger" yaml:"swagger"`
	RefMap              map[string]string
	ApiPathsMap         map[string]map[string]map[string]Path
}

type SecurityDefinition struct {
	Type             string `json:"type"`
	Name             string `json:"name"`
	In               string `json:"in"`
	AuthorizationURL string `json:"authorizationUrl"`
	Flow             string `json:"flow"`
	Scopes           struct {
		ReadPets  string `json:"read:pets"`
		WritePets string `json:"write:pets"`
	} `json:"scopes"`
}

// extractTypeName is a helper function for extracting property type name
// as Go type or custom type name.
func extractTypeName(schema *OpenAPISchema, parentName string, property Property) string {
	res := property.Type
	if property.Format != "" {
		res = property.Format
	}
	if typeName, ok := builtInTypesMap[res]; ok {
		res = typeName
	}
	if refName, ok := schema.RefMap[property.Ref]; ok {
		// return non-required types as pointers
		if refName == parentName {
			res = "*" + strcase.ToCamel(refName)
		} else {
			res = strcase.ToCamel(refName)
		}
	}
	if property.Type != "array" && res != "object" {
		return res
	}
	if property.Items != nil {
		if property.Items.Type != "" {
			res = property.Items.Type
		}
		if refName, ok := schema.RefMap[property.Items.Ref]; ok {
			res = strcase.ToCamel(refName)
		}
	}
	if res != "object" {
		return "[]" + res
	}
	// if property has additional properties then it is a map.
	additionalProperties := property.AdditionalProperties
	mapPrefix := ""
	for additionalProperties != nil {
		mapPrefix += "map[string]"
		res = extractTypeName(schema, parentName, *additionalProperties)
		additionalProperties = additionalProperties.AdditionalProperties
	}
	if mapPrefix != "" {
		return mapPrefix + res
	}
	return "interface{}"
}

var (
	builtInTypesMap = map[string]string{
		"string":    "string",
		"boolean":   "bool",
		"integer":   "int",
		"number":    "float64",
		"date-time": "time.Time",
		"double":    "float64",
		"int64":     "int",
		"int32":     "int",
	}

	builtInGoTypesMap = map[string]string{
		"string":    "string",
		"bool":      " boolean",
		"int":       "integer",
		"float64":   "double",
		"time.Time": "date-time",
	}

	templateFuncs template.FuncMap = template.FuncMap{
		"toCamelCase": func(str string) string {
			return strcase.ToCamel(str)
		},
		"extractTypeName": func(schema *OpenAPISchema, parentName string, property Property) string {
			return extractTypeName(schema, parentName, property)
		},
		// "definitionToProperty": func(definition Definition) Property {
		// 	return Property{
		// 		Description: definition.Description,
		// 		Type:        definition.Type,
		// 		Ref:         definition.Ref,
		// 		Format:      definition.Format,
		// 		XGoName:     definition.XGoName,
		// 	}
		// },
		"toUpperCase": func(str string) string {
			return strings.ToUpper(str)
		},
		"pathParameterToProperty": func(params []PathParameter) Property {
			param := PathParameter{}
			if len(params) > 0 {
				param = params[0]
			}
			return Property{
				Ref: param.Schema.Ref,
			}
		},
		"stringToInt": func(str string) int {
			i, _ := strconv.Atoi(str)
			return i
		},
		"extractResponseTypeFields": func(schema *OpenAPISchema, parentName string, responses map[string]Response) string {
			fields := ""
			// fieldsMap := ""
			// for _, response := range responses {
			// 	// statusCode, _ := strconv.Atoi(statusCodeStr)
			// 	// schemaType := extractTypeName(schema, parentName, response.Schema)

			// }
			return fields
		},
	}
)

// LoadOpenApiSchema loads open api schema from api schema file.
func LoadOpenApiSchema(filePath string) (*OpenAPISchema, error) {
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	schema := OpenAPISchema{
		RefMap:      make(map[string]string),
		ApiPathsMap: make(map[string]map[string]map[string]Path),
	}
	err = yaml.Unmarshal(fileContents, &schema)
	if err != nil {
		return nil, err
	}
	for name := range schema.Definitions {
		key := fmt.Sprintf("#/definitions/%s", name)
		schema.RefMap[key] = name
	}
	for name := range schema.Responses {
		key := fmt.Sprintf("#/responses/%s", name)
		schema.RefMap[key] = name
	}
	// extracting API paths based on path tags.
	for path := range schema.Paths {
		for httpMethod, pathInfo := range schema.Paths[path] {
			for _, tag := range pathInfo.Tags {
				if _, ok := schema.ApiPathsMap[tag]; !ok {
					schema.ApiPathsMap[tag] = make(map[string]map[string]Path)
				}
				if _, ok := schema.ApiPathsMap[tag][path]; !ok {
					schema.ApiPathsMap[tag][path] = make(map[string]Path)
				}
				schema.ApiPathsMap[tag][path][httpMethod] = pathInfo
			}
		}
	}
	return &schema, nil
}

// GenerateGoSDK generates a Go api sdk from an openapi schema file.
func GenerateGoSDK(schemaFile string, outDir string) error {
	err := os.MkdirAll(outDir, 0700)
	if err != nil {
		return err
	}
	schema, err := LoadOpenApiSchema(schemaFile)
	if err != nil {
		return err
	}
	outFile := outDir + "/client.go"
	t, err := template.New("client.go.tmpl").Funcs(templateFuncs).ParseFiles("openapi/templates/client.go.tmpl")
	if err != nil {
		return err
	}
	buffer := &bytes.Buffer{}
	err = t.Execute(buffer, schema)
	if err != nil {
		return err
	}
	processedContents, err := imports.Process(outFile, buffer.Bytes(), nil)
	if err != nil {
		return err
	}
	return os.WriteFile(outFile, processedContents, 0700)
}
