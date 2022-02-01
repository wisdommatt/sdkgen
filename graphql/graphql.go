package graphql

import (
	"bytes"
	"log"
	"os"
	"text/template"

	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
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
	return gqlparser.LoadSchema(sources...)
}

// GenerateSDKClient generates a graphql sdk client from schema.
func GenerateSDKClient(schema *ast.Schema, outFile string) error {
	clientTmp, err := template.ParseFiles("graphql/templates/client.go.tpl")
	if err != nil {
		return err
	}
	buffer := &bytes.Buffer{}
	err = clientTmp.Execute(buffer, schema)
	if err != nil {
		return err
	}
	log.Println(buffer.String())
	return nil
}
