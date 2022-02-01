package graphql

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
)

var (
	graphqlDefaultFieldsMap = map[string]string{
		"Int":     "int",
		"String":  "string",
		"Float":   "float64",
		"ID":      "string",
		"Boolean": "bool",
		"Time":    "time.Time",
		"Date":    "time.Time",
	}

	templateFuncs = template.FuncMap{
		"extractFieldTypeName": func(field *ast.FieldDefinition) string {
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
			if field.Type.NonNull {
				return fieldType
			}
			return "*" + fieldType
		},
	}
)

// LoadGraphqlSchema loads graphql schemas from graphql schema files.
func LoadGraphqlSchema(filenames ...string) (*ast.Schema, error) {
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
	schema, err := gqlparser.LoadSchema(sources...)
	if err != nil {
		return nil, err
	}
	return parseSchema(schema), nil
}

func parseSchema(schema *ast.Schema) *ast.Schema {
	schemaTypes := map[string]*ast.Definition{}
	for key, typ := range schema.Types {
		if _, ok := graphqlDefaultFieldsMap[key]; !ok {
			schemaTypes[key] = typ
		}
	}
	schema.Types = schemaTypes
	return schema
}

// GenerateSDKClient generates a graphql sdk client from schema.
func GenerateSDKClient(schema *ast.Schema, outFile string) error {
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
