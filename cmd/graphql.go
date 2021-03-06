/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/wisdommatt/sdkgen/graphql"
	"github.com/wisdommatt/sdkgen/pkg/log"
)

// graphqlCmd represents the graphql command
var graphqlCmd = &cobra.Command{
	Use:   "graphql",
	Short: "Generate SDK client from graphql schema",
	Run: func(cmd *cobra.Command, args []string) {
		schemaFile, _ := cmd.Flags().GetString("schema")
		if schemaFile == "" {
			log.Fatalln(color.FgRed, "ERROR", "--schema is required")
		}
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			log.Fatalln(color.FgRed, "ERROR", "--output is required")
		}
		err := graphql.GenerateGoSDK(schemaFile, output)
		if err != nil {
			log.Fatalln(color.FgRed, "ERROR", err.Error())
		}
		log.Fatalln(color.FgGreen, "COMPLETED", "API SDK client generated successfully")
	},
}

func init() {
	rootCmd.AddCommand(graphqlCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	graphqlCmd.Flags().String("schema", "", "path to graphql schema file")
	graphqlCmd.Flags().String("output", "", "name/path of generated client package")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// graphqlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
