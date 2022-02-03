package graphql

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"golang.org/x/tools/imports"
)

// Schema contains the data about a graphql schema after
// extraction.
type Schema struct {
	AstSchema     *ast.Schema
	Objects       map[string]*ast.Definition
	Scalars       map[string]*ast.Definition
	Unions        map[string]*ast.Definition
	Enums         map[string]*ast.Definition
	Mutations     []*ast.FieldDefinition
	Queries       []*ast.FieldDefinition
	Subscriptions []*ast.FieldDefinition
}

// NewSchema creates a new schema from an ast schema object.
func NewSchema(astSchema *ast.Schema) *Schema {
	return &Schema{
		AstSchema: astSchema,
		Objects:   make(map[string]*ast.Definition),
		Scalars:   make(map[string]*ast.Definition),
		Unions:    make(map[string]*ast.Definition),
		Enums:     make(map[string]*ast.Definition),
	}
}

var (
	graphqlDefaultFieldsMap = map[string]string{
		"Int":     "int",
		"String":  "string",
		"Float":   "float64",
		"ID":      "string",
		"Boolean": "bool",
		"Time":    "time.Time",
		"Date":    "time.Time",
		"Email":   "string",
	}

	builtInTypesMap = map[string]string{
		"int":     "0",
		"string":  "",
		"float64": "0",
		"bool":    "false",
	}

	templateFuncs = template.FuncMap{
		"extractFieldTypeName": func(schema *Schema, name string, typ *ast.Type) string {
			fieldType := strings.ReplaceAll(typ.Name(), "!", "")

			// checking if field type is an array.
			if typ.Elem != nil {
				if typeName, ok := graphqlDefaultFieldsMap[fieldType]; ok {
					if typ.NonNull {
						return "[]" + typeName
					}
					return "[]*" + typeName
				}
				// return field type as interface{} if initial field type
				// is a graphql scalar.
				if _, ok := schema.Scalars[fieldType]; ok {
					return "[]interface{}"
				}
				if typ.Elem.NonNull {
					return "[]" + strcase.ToCamel(fieldType)
				}
				return "[]*" + strcase.ToCamel(fieldType)
			}

			if typeName, ok := graphqlDefaultFieldsMap[fieldType]; ok {
				if typ.NonNull {
					return typeName
				}
				return "*" + typeName
			}
			// return field type as interface{} if initial field type
			// is a graphql scalar.
			if _, ok := schema.Scalars[fieldType]; ok {
				return "interface{}"
			}
			if typ.NonNull {
				return strcase.ToCamel(fieldType)
			}
			return "*" + strcase.ToCamel(fieldType)
		},
		"toCamelCase": func(str string) string {
			if strings.HasPrefix(str, "__") {
				return str
			}
			return strcase.ToCamel(str)
		},
		"isLastEnumField": func(enum ast.EnumValueList, key int) bool {
			return len(enum)-1 == key
		},
		"isExported": func(str string) bool {
			return !strings.HasPrefix(str, "_")
		},
		"toLowerCamel": func(str string) string {
			return strcase.ToLowerCamel(str)
		},
		"toPointerTypeName": func(schema *Schema, name string, typ *ast.Type) string {
			if _, ok := builtInTypesMap[name]; ok {
				return name
			}
			if _, ok := schema.Unions[name]; ok {
				return name
			}
			if typ.Elem != nil || !typ.NonNull {
				return name
			}
			return "*" + name
		},
		"nilValue": func(typeName string, typ *ast.Type) string {
			if value, ok := builtInTypesMap[typeName]; ok {
				return value
			}
			return "nil"
		},
	}
)

// LoadGraphqlSchema loads graphql schemas from graphql schema files.
func LoadGraphqlSchema(filenames ...string) (*Schema, error) {
	sources := []*ast.Source{}
	for _, filename := range filenames {
		fileContents, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		sources = append(sources, &ast.Source{
			Input: string(fileContents),
		})
	}
	astSchema, err := gqlparser.LoadSchema(sources...)
	if err != nil {
		return nil, err
	}
	schema := NewSchema(astSchema)
	return parseSchema(schema), nil
}

func parseSchema(schema *Schema) *Schema {
	for key, typ := range schema.AstSchema.Types {
		if _, ok := graphqlDefaultFieldsMap[key]; ok {
			continue
		}

		if typ.Name == "Mutation" {
			schema.Mutations = typ.Fields
			continue
		}
		if typ.Name == "Query" {
			schema.Queries = typ.Fields
			continue
		}
		if typ.Name == "Subscription" {
			schema.Subscriptions = typ.Fields
			continue
		}

		switch typ.Kind {
		case ast.Scalar:
			schema.Scalars[key] = typ

		case ast.Union:
			schema.Unions[key] = typ

		case ast.Enum:
			schema.Enums[key] = typ

		case ast.Object, ast.InputObject:
			schema.Objects[key] = typ
		}
	}
	return schema
}

// GenerateSDKClient generates a graphql sdk client from schema.
func GenerateSDKClient(schema *Schema, outFile string) error {
	clientTmp, err := template.New("client.go.tpl").Funcs(templateFuncs).
		ParseFiles("graphql/templates/client.go.tpl")
	if err != nil {
		return err
	}
	buffer := &bytes.Buffer{}
	err = clientTmp.Execute(buffer, schema)
	if err != nil {
		return err
	}
	res, err := imports.Process(outFile, buffer.Bytes(), nil)
	if err != nil {
		return err
	}
	return os.WriteFile(outFile, res, 0666)
}
