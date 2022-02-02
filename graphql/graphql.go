package graphql

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
)

// Schema contains the data about a graphql schema after
// extraction.
type Schema struct {
	AstSchema *ast.Schema
	Objects   map[string]*ast.Definition
	Scalars   map[string]*ast.Definition
	Unions    map[string]*ast.Definition
	Enums     map[string]*ast.Definition
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

	templateFuncs = template.FuncMap{
		"extractFieldTypeName": func(schema *Schema, field *ast.FieldDefinition) string {
			fieldType := strings.ReplaceAll(field.Type.Name(), "!", "")

			// checking if field type is an array.
			if field.Type.Elem != nil {
				if typeName, ok := graphqlDefaultFieldsMap[fieldType]; ok {
					if field.Type.NonNull {
						return "[]" + typeName
					}
					return "[]*" + typeName
				}
				if field.Type.Elem.NonNull {
					return "[]" + fieldType
				}
				return "[]*" + fieldType
			}

			if typeName, ok := graphqlDefaultFieldsMap[fieldType]; ok {
				if field.Type.NonNull {
					return typeName
				}
				return "*" + typeName
			}
			// return field type as interface{} if initial field type
			// is a graphql scalar.
			if _, ok := schema.Scalars[fieldType]; ok {
				return "interface{}"
			}
			if field.Type.NonNull {
				return fieldType
			}
			return "*" + fieldType
		},
		"toCamelCase": func(str string) string {
			return strcase.ToCamel(str)
		},
		"isLastEnumField": func(enum ast.EnumValueList, key int) bool {
			return len(enum)-1 == key
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
	err = os.WriteFile(outFile, buffer.Bytes(), 0600)
	if err != nil {
		return err
	}
	return nil
}
