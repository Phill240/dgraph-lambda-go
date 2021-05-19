package gql

import (
	"fmt"
	"html/template"
	"os"
)

func GenerateGql() error {
	// Generate gqlgen.yml
	f, err := os.Create("gqlgen.yml")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer f.Close()

	var data struct{}
	gqlConfigTemplate.Execute(f, data)

	// Generate gqlgen.yml
	os.Mkdir("graph", os.ModePerm)

	f, err = os.Create("graph/schema.graphqls")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer f.Close()

	schemaTemplate.Execute(f, data)

	return nil
}

var gqlConfigTemplate = template.Must(template.New("gql-config").Parse(`
# Where are all the schema files located? globs are supported eg  src/**/*.graphqls, will support loading from web
schema:
  - graph/*.graphqls

# Where should the generated server code go?
exec:
  filename: graph/generated/generated.go
  package: generated

# Uncomment to enable federation
# federation:
#   filename: graph/generated/federation.go
#   package: generated

# Where should any generated models go?
model:
  filename: graph/model/models_gen.go
  package: model

# Where should the resolver implementations go?
resolver:
  layout: follow-schema
  dir: graph
  package: graph

# gqlgen will search for any type names in the schema in these go packages
# if they match it will use them, otherwise it will generate them.
autobind:
  - "github.com/schartey/dgraph-lambda-go/graph/model"

models:
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
      - github.com/99designs/gqlgen/graphql.Int32
  Int:
    model:
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
      - github.com/99designs/gqlgen/graphql.Int32

  DateTime:
    model:
      - github.com/schartey/dgraph-lambda-go/dgraph.DateTime
`))

var schemaTemplate = template.Must(template.New("schema").Parse(`type User {
	id: ID!
	username: String!
	firstname: String!
	lastname: String!
	name: String! @lambda
}

input CreateUserInput {
	username: String!
}

type Mutation {
    createUser(input: CreateUserInput!): User @lambda
}
`))
