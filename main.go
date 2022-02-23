package main

import (
	"log"

	"github.com/wisdommatt/sdkgen/graphql"
	"github.com/wisdommatt/sdkgen/openapi"
)

func main() {
	err := graphql.GenerateGoSDK("sample.graphql", "sample-gen")
	if err != nil {
		log.Fatal(err)
	}
	err = openapi.GenerateGoSDK("openapi-sample.yaml", "sample-gen/openapii")
	if err != nil {
		log.Fatal(err)
	}
}
