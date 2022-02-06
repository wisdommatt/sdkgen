package openapi

import (
	"bytes"
	"html/template"
	"os"

	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v2"
)

type DefinitionProperty struct {
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Example     interface{} `json:"example"`
	XGoName     string      `json:"x-go-name"`
	Ref         string      `json:"$ref"`
}

type Definition struct {
	Description string                        `json:"description"`
	Properties  map[string]DefinitionProperty `json:"properties"`
	Required    []string                      `json:"required"`
	Type        string                        `json:"type"`
	XGoPackage  string                        `json:"x-go-package"`
}

type Info struct {
	Description string `json:"description"`
	Title       string `json:"title"`
	Version     string `json:"version"`
}

type Path struct {
	Description string              `json:"description"`
	OperationID string              `json:"operationId"`
	Parameters  []PathParameter     `json:"parameters"`
	Responses   map[int]RefSchema   `json:"responses"`
	Summary     string              `json:"summary"`
	Tags        []string            `json:"tags"`
	Security    map[string][]string `json:"security"`
}

type PathParameter struct {
	In       string    `json:"in"`
	Name     string    `json:"name"`
	Required bool      `json:"required"`
	Schema   RefSchema `json:"schema"`
}

type RefSchema struct {
	Ref string `json:"$ref"`
}

type Response struct {
	Description string         `json:"description"`
	Schema      ResponseSchema `json:"schema"`
}

type ResponseSchema struct {
	Properties map[string]ResponseSchemaProperty `json:"properties"`
	Required   []string                          `json:"required"`
	Type       string                            `json:"type"`
	Ref        string                            `json:"$ref"`
}

type ResponseSchemaProperty struct {
	Description string `json:"description"`
	Example     string `json:"example"`
	Type        string `json:"type"`
	XGoName     string `json:"x-go-name"`
}

type SecurityDefinition struct {
	Type string `json:"type"`
}

type OpenAPISchema struct {
	BasePath            string                        `json:"basePath"`
	Consumes            []string                      `json:"consumes"`
	Definitions         map[string]Definition         `json:"definitions"`
	Host                string                        `json:"host"`
	Info                Info                          `json:"info"`
	Paths               map[string]Path               `json:"paths"`
	Produces            []string                      `json:"produces"`
	Responses           map[string]Response           `json:"responses"`
	Schemes             []string                      `json:"schemes"`
	SecurityDefinitions map[string]SecurityDefinition `json:"securityDefinitions"`
	Swagger             string                        `json:"swagger"`
}

var (
	builtInTypesMap = map[string]string{
		"string":  "string",
		"boolean": "bool",
		"integer": "int",
		"number":  "float64",
	}

	templateFuncs template.FuncMap = template.FuncMap{
		"toCamelCase": func(str string) string {
			return strcase.ToCamel(str)
		},
		"extractTypeName": func(name, ref string) string {
			if typeName, ok := builtInTypesMap[name]; ok {
				return typeName
			}
			return "interface{}"
		},
	}
)

// LoadOpenApiSchema loads open api schema from api schema file.
func LoadOpenApiSchema(filePath string) (*OpenAPISchema, error) {
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var schema OpenAPISchema
	err = yaml.Unmarshal(fileContents, &schema)
	if err != nil {
		return nil, err
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
	return os.WriteFile(outFile, buffer.Bytes(), 0700)
}
