package codegen

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/schartey/dgraph-lambda-go/codegen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

type model struct {
	Name   string
	Fields []*field
}

type field struct {
	Name   string
	GoType *graphql.GoTypeDefinition
}

type ModelGenerator struct {
	schema   *ast.Schema
	allTypes map[string]*ast.Definition
	models   map[string]*model
}

func NewModelGenerator(schema *ast.Schema) *ModelGenerator {
	models := make(map[string]*model)

	return &ModelGenerator{schema: schema, models: models}
}

func (m *ModelGenerator) Parse() {
	m.collectAllTypes()

	for _, typ := range m.schema.Types {
		if !typ.BuiltIn {
			model, err := m.parseType(typ, false)
			if err == nil {
				m.models[typ.Name] = model
			}
		}
	}
}

func (m *ModelGenerator) Generate() error {
	f, err := os.Create("generated/models_gen.go")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer f.Close()

	pkgs := make(map[string]bool)

	for _, model := range m.models {
		for _, field := range model.Fields {
			if field.GoType.PkgName != "" {
				pkgs[field.GoType.PkgName] = true
			}
		}
	}

	return modelTemplate.Execute(f, struct {
		Models   map[string]*model
		Packages map[string]bool
	}{
		Models:   m.models,
		Packages: pkgs,
	})
}

func (m *ModelGenerator) collectAllTypes() {
	m.allTypes = make(map[string]*ast.Definition)
	for _, typ := range m.schema.Types {
		if !typ.BuiltIn {
			m.allTypes[typ.Name] = typ
		}
	}
}

func (m *ModelGenerator) parseType(typeDef *ast.Definition, force bool) (*model, error) {
	if force || m.schemaTypeContainsLambdaField(typeDef) {
		fmt.Println(typeDef.Name)
		var fields []*field
		for _, field := range typeDef.Fields {
			generatedField, err := m.parseField(field)
			if err == nil {
				fields = append(fields, generatedField)
			}
		}
		return &model{Name: typeDef.Name, Fields: fields}, nil
	}
	return nil, errors.New("Type has no lambda field")
}

func (m *ModelGenerator) parseField(typeDef *ast.FieldDefinition) (*field, error) {
	// Non Scalar
	if !graphql.IsDgraphType(typeDef.Type.Name()) && m.allTypes[typeDef.Type.Name()] != nil {
		// Generate type before continuing
		if m.models[typeDef.Type.Name()] == nil {
			// Reserve current type
			m.models[typeDef.Type.Name()] = &model{}
			// Generate current type
			parsedType, err := m.parseType(m.allTypes[typeDef.Type.Name()], true)
			if err == nil {
				m.models[typeDef.Type.Name()] = parsedType
			}
		}

		if typeDef.Directives.ForName("hasInverse") == nil {
			var typeName string

			if graphql.IsArray(typeDef.Type.String()) {
				typeName = fmt.Sprintf("[]%s", typeDef.Type.Name())
			} else {
				typeName = typeDef.Type.Name()
			}
			return &field{Name: typeDef.Name, GoType: &graphql.GoTypeDefinition{TypeName: typeName}}, nil
		}
	} else {
		goType, err := graphql.SchemaTypeToGoType(typeDef.Name, typeDef.Type)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Field %s cannot be generated", typeDef.Name))
		}
		return &field{Name: typeDef.Name, GoType: goType}, nil
	}

	return nil, errors.New(fmt.Sprintf("Field %s cannot be generated", typeDef.Name))
}

func (m *ModelGenerator) schemaTypeContainsLambdaField(schemaType *ast.Definition) bool {
	for _, field := range schemaType.Fields {
		if field.Directives.ForName("lambda") != nil {
			return true
		}
	}
	return false
}

var modelTemplate = template.Must(template.New("model").Funcs(template.FuncMap{
	"title": strings.Title,
}).Parse(`package generated
{{ if .Packages }}
import({{range $pkg, $b := .Packages}}
	"{{$pkg}}"{{end}}
) {{end}}
{{range $name, $model := .Models}}
type {{$model.Name}} struct { {{range $field := $model.Fields}}
	{{$field.Name | title }} {{$field.GoType.TypeName}} ` + "`json:\"{{$field.Name}}\"`" + `{{end}}
}
{{end}}`))
