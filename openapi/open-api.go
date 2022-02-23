package openapi

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

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
		Name    string `json:"name" yaml:"name"`
		Wrapped bool   `json:"wrapped" yaml:"wrapped"`
	} `json:"xml"`
	Default struct {
		Description string `json:"description" yaml:"description"`
	} `json:"default"`
	Schema *Property `json:"schema" yaml:"schema"`
}

func (p Property) IsRequired(str string) bool {
	for _, r := range p.Required {
		if r == str {
			return true
		}
	}
	return false
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
	Produces    []string              `json:"produces" yaml:"produces"`
}

type PathParameter struct {
	Description string   `json:"description" yaml:"description"`
	In          string   `json:"in" yaml:"in"`
	Name        string   `json:"name" yaml:"name"`
	Required    bool     `json:"required" yaml:"required"`
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
	RefPropertyMap      map[string]Property
	ApiPathsMap         map[string]map[string]map[string]Path
}

type SecurityDefinition struct {
	Type             string `json:"type" yaml:"type"`
	Name             string `json:"name" yaml:"name"`
	In               string `json:"in" yaml:"in"`
	AuthorizationURL string `json:"authorizationUrl" yaml:"authorizationUrl"`
	Flow             string `json:"flow" yaml:"flow"`
	Scopes           struct {
		ReadPets  string `json:"read:pets" yaml:"read:pets"`
		WritePets string `json:"write:pets" yaml:"write:pets"`
	} `json:"scopes"`
}

var (
	//go:embed templates/client.go.tmpl
	clientTemplateFile string

	builtInTypesMap = map[string]string{
		"string":      "string",
		"boolean":     "bool",
		"integer":     "int",
		"number":      "float64",
		"date-time":   "time.Time",
		"double":      "float64",
		"int64":       "int",
		"int32":       "int",
		"int":         "int",
		"float64":     "float64",
		"interface{}": "interface{}",
		"time.Time":   "time.Time",
	}

	templateFuncs template.FuncMap = template.FuncMap{
		"toCamelCase": func(str string) string {
			return strcase.ToCamel(str)
		},
		"extractTypeName": func(schema *OpenAPISchema, property Property) TypeName {
			return extractTypeName(schema, property)
		},
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
		"extractResponseType": func(schema *OpenAPISchema, responseName string, responses map[string]Property) string {
			fieldsMap := map[string]string{}
			for _, response := range responses {
				definition := extractRootDefinition(schema, response.Ref)
				if definition == nil {
					continue
				}
				for name, prop := range definition.Properties {
					fieldType := extractTypeName(schema, prop)
					prefix := ""
					if !definition.IsRequired(name) && !fieldType.IsNullable() && !fieldType.IsBuiltIn() {
						prefix = "*"
					}
					existingField, ok := fieldsMap[name]
					if existingField == "interface{}" || !ok || existingField == fieldType.String() {
						fieldsMap[name] = prefix + fieldType.String()
						continue
					}
					fieldsMap[name] = "interface{}"
				}
			}
			if len(fieldsMap) == 0 {
				return "interface{}"
			}
			responseType := "struct { \n"
			for fieldName, fieldType := range fieldsMap {
				responseType += fmt.Sprintf(
					"%s %s `json:\"%s,omitempty\"` \n",
					strcase.ToCamel(fieldName),
					fieldType,
					fieldName,
				)
			}
			return responseType + "}"
		},
		"pointerPrefix": func(property Property, fieldName string, typeName TypeName) TypeName {
			if !property.IsRequired(fieldName) && !typeName.IsBuiltIn() && !typeName.IsNullable() {
				return "*" + typeName
			}
			return typeName
		},
	}
)

type TypeName string

func (t TypeName) IsNullable() bool {
	str := t.String()
	return strings.Contains(str, "[") || strings.Contains(str, "*")
}

func (t TypeName) IsBuiltIn() bool {
	_, ok := builtInTypesMap[string(t)]
	return ok
}

func (t TypeName) String() string {
	return string(t)
}

// extractTypeName is a helper function for extracting property type name
// as Go type or custom type name.
func extractTypeName(schema *OpenAPISchema, property Property) TypeName {
	res := property.Type
	if property.Format != "" {
		res = property.Format
	}
	if typeName, ok := builtInTypesMap[res]; ok {
		res = typeName
	}
	if refName, ok := schema.RefMap[property.Ref]; ok {
		// return non-required types as pointers
		res = strcase.ToCamel(refName)
	}
	if property.Type != "array" && res != "object" {
		return TypeName(res)
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
		return TypeName("[]" + res)
	}
	// if property has additional properties then it is a map.
	additionalProperties := property.AdditionalProperties
	mapPrefix := ""
	for additionalProperties != nil {
		mapPrefix += "map[string]"
		typeName := extractTypeName(schema, *additionalProperties)
		res = typeName.String()
		additionalProperties = additionalProperties.AdditionalProperties
	}
	if mapPrefix != "" {
		return TypeName(mapPrefix + res)
	}
	return "interface{}"
}

func extractRootDefinition(schema *OpenAPISchema, ref string) *Property {
	definition, ok := schema.RefPropertyMap[ref]
	if !ok {
		return nil
	}
	if strings.Contains(ref, "definitions") {
		return &definition
	}
	if definition.Schema == nil || definition.Schema.Ref == "" {
		return definition.Schema
	}
	return extractRootDefinition(schema, definition.Schema.Ref)
}

// LoadOpenApiSchema loads open api schema from api schema file.
func LoadOpenApiSchema(filePath string) (*OpenAPISchema, error) {
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	schema := OpenAPISchema{
		RefMap:         make(map[string]string),
		RefPropertyMap: make(map[string]Property),
		ApiPathsMap:    make(map[string]map[string]map[string]Path),
	}
	if strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml") {
		err = yaml.Unmarshal(fileContents, &schema)
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(filePath, ".json") {
		err = json.Unmarshal(fileContents, &schema)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("provide a valid json / yaml schema file")
	}
	for name, property := range schema.Definitions {
		key := fmt.Sprintf("#/definitions/%s", name)
		schema.RefMap[key] = name
		schema.RefPropertyMap[key] = property
	}
	for name, property := range schema.Responses {
		key := fmt.Sprintf("#/responses/%s", name)
		schema.RefMap[key] = name
		schema.RefPropertyMap[key] = property
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
	t, err := template.New("client.go.tmpl").Funcs(templateFuncs).Parse(clientTemplateFile)
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
