package main

import (
	"log"

	"github.com/wisdommatt/sdkgen/graphql"
)

func main() {
	schema, _ := graphql.LoadGraphqlSchema("sample.graphql")
	// err := graphql.GenerateSDKClient(schema, "sdks/example.go")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	for _, typ := range schema.Types {
		if typ.Name == "UserCategory" {
			log.Println(typ.Name, typ.Kind, typ.BuiltIn, typ.Description, typ.Directives, typ.Fields)
			// for _, field := range typ.Fields {
			// 	log.Println(field.Name, field.Type, field.Arguments, field.Type.Elem, field.Type.NonNull)

			// 	for _, argument := range field.Arguments {
			// 		log.Println(argument.Name, argument.Type)
			// 	}
			// }
		}
	}
}
