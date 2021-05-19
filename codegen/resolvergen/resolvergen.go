package resolvergen

import (
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/gqlgen/codegen"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/99designs/gqlgen/plugin"
	"github.com/schartey/dgraph-lambda-go/dgraph"
	"github.com/vektah/gqlparser/v2/ast"
)

func NewResolverPlugin() *ResolverPlugin {
	return &ResolverPlugin{}
}

type ResolverPlugin struct{}

var _ plugin.CodeGenerator = &ResolverPlugin{}

func (m *ResolverPlugin) Name() string {
	return "resolvergen"
}

func (r *ResolverPlugin) MutateConfig(cfg *config.Config) error {
	fmt.Println("Generating Resolvers")

	cfg.Directives["hasInverse"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["search"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["dgraph"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["id"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["withSubscription"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["secret"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["auth"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["custom"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["remote"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["remoteResponse"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["cascade"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["lambda"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["lambdaOnMutate"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["cacheControl"] = config.DirectiveConfig{SkipRuntime: true}
	cfg.Directives["generate"] = config.DirectiveConfig{SkipRuntime: true}

	return nil
}

func (m *ResolverPlugin) GenerateCode(data *codegen.Data) error {

	// These two generate the Resolver type that contains the resolver functions to be called
	// Filter lambda queries
	if data.QueryRoot != nil {
		var queryFields []*codegen.Field
		for _, f := range data.QueryRoot.Fields {
			if f.IsResolver {
				if isLambdaField(f) {
					queryFields = append(queryFields, f)
				}
			} else {
				queryFields = append(queryFields, f)
			}
		}
		data.QueryRoot.Fields = queryFields
	}

	// Filter lambda mutations
	if data.MutationRoot != nil {
		var mutationFields []*codegen.Field
		for _, f := range data.MutationRoot.Fields {
			if f.IsResolver {
				if isLambdaField(f) {
					mutationFields = append(mutationFields, f)
				}
			} else {
				mutationFields = append(mutationFields, f)
			}
		}
		data.MutationRoot.Fields = mutationFields
	}

	// Here we filter the models that actually need to be generated

	// This generates the actual functions that are called by the Resolver
	// Filter generated resolvers
	files := map[string]*File{}

	for _, o := range data.Objects {
		if !o.BuiltIn && !o.HasResolvers() {
		}
		if o.HasResolvers() {
			fn := gqlToResolverName(data.Config.Resolver.Dir(), o.Position.Src.Name, data.Config.Resolver.FilenameTemplate)
			if files[fn] == nil {
				files[fn] = &File{}
			}

			for _, f := range o.Fields {
				if !f.IsResolver {
					continue
				}

				if isLambdaField(f) {
					implementation := ""
					if implementation == "" {
						implementation = `panic(fmt.Errorf("not implemented"))`
					}

					resolver := Resolver{o, f, implementation}
					fn := gqlToResolverName(data.Config.Resolver.Dir(), f.Position.Src.Name, data.Config.Resolver.FilenameTemplate)
					if files[fn] == nil {
						files[fn] = &File{}
					}

					files[fn].Resolvers = append(files[fn].Resolvers, &resolver)
				}
			}
			files[fn].Objects = append(files[fn].Objects, o)
		}

		for filename, file := range files {
			resolverBuild := &ResolverBuild{
				File:         file,
				PackageName:  data.Config.Resolver.Package,
				ResolverType: data.Config.Resolver.Type,
			}

			err := templates.Render(templates.Options{
				PackageName: data.Config.Resolver.Package,
				FileNotice: `
					// This file will be automatically regenerated based on the schema, any resolver implementations
					// will be copied through when generating and any unknown code will be moved to the end.`,
				Filename: filename,
				Data:     resolverBuild,
				Packages: data.Config.Packages,
			})
			if err != nil {
				return err
			}
		}
	}

	if _, err := os.Stat(data.Config.Resolver.Filename); os.IsNotExist(err) {
		err := templates.Render(templates.Options{
			PackageName: data.Config.Resolver.Package,
			FileNotice: `
				// This file will not be regenerated automatically.
				//
				// It serves as dependency injection for your app, add any dependencies you require here.`,
			Template: `type {{.}} struct {}`,
			Filename: data.Config.Resolver.Filename,
			Data:     data.Config.Resolver.Type,
			Packages: data.Config.Packages,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: Add all needed directives and types
func (m *ResolverPlugin) InjectSourceEarly() *ast.Source {
	return &ast.Source{
		Name:    "dgraph/lambda.graphql",
		Input:   dgraph.SchemaInputs + dgraph.DirectiveDefs + dgraph.FilterInputs,
		BuiltIn: true,
	}
}

func schemaTypeContainsLambdaField(schemaType *ast.Definition) bool {
	for _, field := range schemaType.Fields {
		if field.Directives.ForName("lambda") != nil {
			return true
		}
	}
	return false
}

func isLambdaField(field *codegen.Field) bool {
	isLambda := false
	for _, d := range field.Directives {
		if d.Name == "lambda" {
			isLambda = true
		}
	}
	return isLambda
}

func gqlToResolverName(base string, gqlname, filenameTmpl string) string {
	gqlname = filepath.Base(gqlname)
	ext := filepath.Ext(gqlname)
	if filenameTmpl == "" {
		filenameTmpl = "{name}.resolvers.go"
	}
	filename := strings.ReplaceAll(filenameTmpl, "{name}", strings.TrimSuffix(gqlname, ext))
	return filepath.Join(base, filename)
}

func isStruct(t types.Type) bool {
	_, is := t.Underlying().(*types.Struct)
	return is
}

type ResolverBuild struct {
	*File
	HasRoot      bool
	PackageName  string
	ResolverType string
}

type File struct {
	// These are separated because the type definition of the resolver object may live in a different file from the
	//resolver method implementations, for example when extending a type in a different graphql schema file
	Objects         []*codegen.Object
	Resolvers       []*Resolver
	RemainingSource string
}

type Resolver struct {
	Object         *codegen.Object
	Field          *codegen.Field
	Implementation string
}
