package graphql

import (
	"os"

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
