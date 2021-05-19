package cmd

import (
	"fmt"
	"os"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/schartey/dgraph-lambda-go/codegen/gql"
	"github.com/schartey/dgraph-lambda-go/codegen/resolvergen"
	"github.com/schartey/dgraph-lambda-go/codegen/servergen"
	"github.com/urfave/cli/v2"
)

var initCmd = &cli.Command{
	Name:  "init",
	Usage: "generate a basic server",
	Action: func(ctx *cli.Context) error {
		err := gql.GenerateGql()
		if err != nil {
			return err
		}

		cfg, err := config.LoadConfigFromDefaultLocations()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to load config", err.Error())
			os.Exit(2)
		}

		// Generate models and resolvers
		err = api.Generate(cfg,
			api.AddPlugin(resolvergen.NewResolverPlugin()),
			api.AddPlugin(servergen.NewServerPlugin("server.go")), // This is the magic line
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return err
		}
		return nil
	},
}
