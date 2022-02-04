package main

import (
	"log"

	"github.com/wisdommatt/sdkgen/graphql"
)

func main() {
	err := graphql.GenerateGoSDK("sample.graphql", "sample-gen")
	if err != nil {
		log.Fatal(err)
	}
}
