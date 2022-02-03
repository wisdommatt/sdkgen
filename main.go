package main

import (
	"log"

	"github.com/wisdommatt/sdkgen/graphql"
)

func main() {
	schema, _ := graphql.LoadGraphqlSchema("sample.graphql")
	err := graphql.GenerateSDKClient(schema, "sample-gen/example.go")
	if err != nil {
		log.Fatal(err)
	}
}
