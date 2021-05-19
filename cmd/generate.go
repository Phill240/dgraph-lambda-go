package cmd

import (
	"fmt"
	"os"

	"github.com/schartey/dgraph-lambda-go/codegen/resolvergen"
	"github.com/urfave/cli/v2"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
)

var generateCmd = &cli.Command{
	Name:  "generate",
	Usage: "generate resolvers and types from schema",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "schema, s", Usage: "the schema filename"},
	},
	Action: func(ctx *cli.Context) error {
		cfg, err := config.LoadConfigFromDefaultLocations()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to load config", err.Error())
			os.Exit(2)
		}

		err = api.Generate(cfg,
			api.AddPlugin(resolvergen.NewResolverPlugin()), // This is the magic line
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return err
		}
		return nil
	},
}
