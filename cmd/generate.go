package cmd

import (
	"fmt"
	"os"

	"github.com/schartey/dgraph-lambda-go/codegen"
	"github.com/schartey/dgraph-lambda-go/codegen/graphql"
	"github.com/urfave/cli/v2"
)

var generateCmd = &cli.Command{
	Name:  "generate",
	Usage: "generate resolvers and types from schema",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "schema, s", Usage: "the schema filename"},
	},
	Action: func(ctx *cli.Context) error {
		schemaFile := ctx.String("schema")

		schema, err := graphql.SchemaLoaderFromFile(schemaFile)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		_, err = os.ReadDir("generated")
		if err != nil {
			if os.IsNotExist(err) {
				os.Mkdir("generated", os.ModePerm)
			}
		}

		modelGenerator := codegen.NewModelGenerator(schema)

		modelGenerator.Parse()
		err = modelGenerator.Generate()
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return nil
	},
}
