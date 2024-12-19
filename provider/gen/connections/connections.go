package connections

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"unicode"

	"github.com/AlekSi/pointer"
	"github.com/invopop/jsonschema"
	"github.com/polytomic/polytomic-go"
	ptclient "github.com/polytomic/polytomic-go/client"
	"github.com/polytomic/polytomic-go/option"
)

const (
	// General
	outputPath      = "./provider"
	exportTemplate  = "./provider/gen/connections/connections.go.tmpl"
	connectionTypes = "./provider/gen/connections/connectiontypes.json"
	jsonschemaPath  = "./provider/gen/connections/connectiontypes"

	// Resources
	connectionResourceTemplate = "./provider/gen/connections/resource.go.tmpl"
	exampleResourceTemplate    = "./provider/gen/connections/resource.tf.go.tmpl"
	exampleResourceOutputPath  = "./examples/resources"

	// Datasources
	connectionDataSourceTemplate = "./provider/gen/connections/datasource.go.tmpl"
	exampleDatasourceTemplate    = "./provider/gen/connections/datasource.tf.go.tmpl"
	exampleDatasourceOutputPath  = "./examples/data-sources"
)

var (
	TypeMap = map[string]Typer{
		"array": {
			AttrType:      "schema.StringAttribute",
			TfType:        "types.String",
			NewAttrType:   "types.StringType",
			Default:       "stringdefault.StaticString(\"\")",
			DefaultImport: "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault",
			GoType:        "string",
		},
		"object": {
			AttrType:      "schema.StringAttribute",
			TfType:        "types.String",
			NewAttrType:   "types.StringType",
			Default:       "stringdefault.StaticString(\"\")",
			DefaultImport: "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault",
			GoType:        "string",
		},

		"string": {
			AttrType:      "schema.StringAttribute",
			TfType:        "types.String",
			NewAttrType:   "types.StringType",
			Default:       "stringdefault.StaticString(\"\")",
			DefaultImport: "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault",
			GoType:        "string",
		},
		"number": {
			AttrType:      "schema.NumberAttribute",
			TfType:        "types.Number",
			NewAttrType:   "types.NumberType",
			Default:       "int64default.StaticInt64(0)",
			DefaultImport: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
			GoType:        "int64",
		},
		"bool": {
			AttrType:    "schema.BoolAttribute",
			TfType:      "types.Bool",
			NewAttrType: "types.BoolType",
			GoType:      "bool",
		},
		"boolean": {
			AttrType:    "schema.BoolAttribute",
			TfType:      "types.Bool",
			NewAttrType: "types.BoolType",
			GoType:      "bool",
		},
		"int": {
			AttrType:      "schema.Int64Attribute",
			TfType:        "types.Int64",
			NewAttrType:   "types.NumberType",
			Default:       "int64default.StaticInt64(0)",
			DefaultImport: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
			GoType:        "int64",
		},
		"int64": {
			AttrType:      "schema.Int64Attribute",
			TfType:        "types.Int64",
			NewAttrType:   "types.NumberType",
			Default:       "int64default.StaticInt64(0)",
			DefaultImport: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
			GoType:        "int64",
		},
		"integer": {
			AttrType:      "schema.Int64Attribute",
			TfType:        "types.Int64",
			NewAttrType:   "types.NumberType",
			Default:       "int64default.StaticInt64(0)",
			DefaultImport: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
			GoType:        "int64",
		},
	}
)

type Typer struct {
	AttrType      string
	TfType        string
	NewAttrType   string
	Default       string
	DefaultImport string
	GoType        string
}

type Connections struct {
	Connections []Connection `yaml:"connections"`
}

type Connection struct {
	// Name is the human readable name for the connection type
	Name string `yaml:"name"`
	// Conn is the connection type name formatted for use in the Terraform
	// resource.
	Conn string `yaml:"-"`
	// Connection is the connection type name formatted for use in the Terraform
	// resource.
	Connection string `yaml:"connection"`
	// ResourceName overrides the name of the resource; if not present the
	// connection type is used.
	ResourceName string
	// Type is the Polytomic connection type.
	Type         string          `yaml:"type"`
	Attributes   []Attribute     `yaml:"attributes"`
	Config       string          `yaml:"config"`
	Datasource   bool            `yaml:"datasource"`
	Resource     bool            `yaml:"resource"`
	ExtraImports map[string]bool `yaml:"-"`
	Imports      string          `yaml:"-"`
}

type Attribute struct {
	Name                string `yaml:"name"`
	CapName             string `yaml:"-"`
	NameOverride        string `yaml:"name_override"`
	Alias               string `yaml:"alias"`
	Sensitive           bool   `yaml:"sensitive"`
	Required            bool   `yaml:"required"`
	Optional            bool   `yaml:"optional"`
	Computed            bool   `yaml:"computed"`
	Type                string `yaml:"type"`
	Description         string `yaml:"description"`
	Example             string `yaml:"example"`
	ExampleTypeOverride string `yaml:"example_type_override"`

	TfType      string `yaml:"-"`
	AttrType    string `yaml:"-"`
	NewAttrType string `yaml:"-"`
	AttrName    string `yaml:"-"`
	Default     string `yaml:"-"`
}

var defaultImports = `
"context"
"errors"
"fmt"
"net/http"
"strings"

"github.com/mitchellh/mapstructure"
"github.com/AlekSi/pointer"
"github.com/mitchellh/mapstructure"
"github.com/hashicorp/terraform-plugin-framework/attr"
"github.com/hashicorp/terraform-plugin-framework/path"
"github.com/hashicorp/terraform-plugin-framework/resource"
"github.com/hashicorp/terraform-plugin-framework/resource/schema"
"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
"github.com/hashicorp/terraform-plugin-framework/types"
"github.com/hashicorp/terraform-plugin-log/tflog"
"github.com/polytomic/polytomic-go"
ptcore "github.com/polytomic/polytomic-go/core"
"github.com/polytomic/terraform-provider-polytomic/provider/internal/client"
`

type Importable struct {
	Name         string
	ResourceName string
}

func fetchOrRead[T any, PT *T](ctx context.Context, path string, fetch func(context.Context) (PT, error)) (PT, error) {
	data, err := fetch(ctx)
	if err != nil {
		if ct, err := os.Open(path); err == nil {
			err = json.NewDecoder(ct).Decode(&data)
			if err != nil {
				return nil, fmt.Errorf("error reading %s: %w", path, err)
			}
		}
		if data == nil {
			// couldn't fetch or read
			return nil, err
		}
	} else {
		// write the fetched data to path
		f, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("error creating %s: %w", path, err)
		}
		defer f.Close()
		err = json.NewEncoder(f).Encode(data)
		if err != nil {
			return nil, fmt.Errorf("error encoding %s: %w", path, err)
		}
	}
	return data, nil
}

func GenerateConnections(ctx context.Context) error {
	client := ptclient.NewClient(
		option.WithBaseURL(os.Getenv("POLYTOMIC_DEPLOYMENT_URL")),
		option.WithToken(os.Getenv("POLYTOMIC_API_KEY")),
	)
	data, err := fetchOrRead(ctx,
		connectionTypes,
		func(ctx context.Context) (*polytomic.ConnectionTypeResponseEnvelope, error) {
			return client.Connections.GetTypes(ctx)
		},
	)

	resources := []Importable{}
	datasources := []Importable{}

	for _, connType := range data.Data {
		connSchema, err := fetchOrRead(ctx,
			filepath.Join(jsonschemaPath, fmt.Sprintf("%s.json", pointer.Get(connType.Id))),
			func(ctx context.Context) (*polytomic.JsonschemaSchema, error) {
				return client.Connections.GetConnectionTypeSchema(ctx, pointer.Get(connType.Id))
			},
		)
		if err != nil {
			return err
		}
		r := Connection{
			Name:         cmp.Or(pointer.Get(connType.Name), pointer.Get(connType.Id)),
			ResourceName: pointer.Get(connType.Id),
			Connection:   pointer.Get(connType.Id),
			Type:         pointer.Get(connType.Id),
			Datasource:   true,
		}

		r.ExtraImports = make(map[string]bool)
		for _, k := range slices.Sorted(maps.Keys(*connSchema.Properties)) {
			p := (*connSchema.Properties)[k]
			a := jsonschema.Schema{}
			propJSON, _ := json.Marshal(p)
			err := json.Unmarshal(propJSON, &a)
			if err != nil {
				return err
			}
			t, ok := TypeMap[a.Type]
			if !ok {
				return fmt.Errorf("type %s not found for %s", a.Type, r.Name)
			}
			var ex string
			if len(a.Examples) > 0 {
				if exstr, ok := a.Examples[0].(string); ok {
					ex = exstr
				}
			}
			attr := Attribute{
				TfType:      t.TfType,
				AttrType:    t.AttrType,
				NewAttrType: t.NewAttrType,
				AttrName:    ValidName(k), // key in the tf schema?
				CapName:     strings.Title(k),
				Name:        k, // key in the payload
				Type:        t.GoType,
				Description: a.Description,
				Example:     ex,
				Sensitive:   p.(map[string]interface{})["sensitive"] == true,
			}
			if a.Format == "json" && attr.Example != "" {
				attr.Example = fmt.Sprintf("jsonencode(%s)", attr.Example)
				attr.ExampleTypeOverride = "json"
			}
			// if a.NameOverride != "" {
			// 	r.Attributes[i].AttrName = a.NameOverride
			// }
			attr.Computed = a.ReadOnly
			attr.Required = slices.Contains(connSchema.Required, k)
			attr.Optional = !attr.Required && !attr.Computed
			if attr.Computed {
				attr.Default = t.Default
				if t.DefaultImport != "" {
					r.ExtraImports[t.DefaultImport] = true
				}
			}
			r.Attributes = append(r.Attributes, attr)
		}
		r.Resource = len(r.Attributes) > 0
		if r.Name == "" {
			r.Name = strings.Title(r.Connection)
		}
		if r.Resource {
			err := writeConnectionResource(r)
			if err != nil {
				return err
			}
			i := Importable{
				Name:         r.Connection,
				ResourceName: fmt.Sprintf("%sConnectionResource", strings.Title(r.Connection)),
			}
			if r.Type != "" {
				i.Name = r.Type
			}
			resources = append(resources, i)
		}
		if r.Datasource {
			err := writeConnectionDataSource(r)
			if err != nil {
				return err
			}
			datasources = append(datasources, Importable{
				Name:         r.Connection,
				ResourceName: fmt.Sprintf("%sConnectionDataSource", strings.Title(r.Connection)),
			})
		}

		err = writeConnectionExamples(r)
		if err != nil {
			return err
		}

	}

	err = writeExports(datasources, resources)
	if err != nil {
		return err
	}

	return nil
}

func writeConnectionExamples(r Connection) error {
	var attributes []Attribute
	for _, a := range r.Attributes {
		if a.ExampleTypeOverride != "" {
			a.Type = a.ExampleTypeOverride
		}
		attributes = append(attributes, a)
	}

	if r.Resource {
		tmpl, err := template.New("resource.tf.go.tmpl").ParseFiles(exampleResourceTemplate)
		if err != nil {
			return err
		}
		newpath := filepath.Join(
			exampleResourceOutputPath,
			fmt.Sprintf("polytomic_%s_connection", r.Connection),
		)
		err = os.MkdirAll(newpath, os.ModePerm)
		if err != nil {
			return err
		}
		f, err := os.Create(
			filepath.Join(newpath, "resource.tf"))

		if err != nil {
			return err
		}
		defer f.Close()

		err = tmpl.Execute(f, struct {
			Resource   string
			Name       string
			Attributes []Attribute
		}{
			Resource:   TerraformResourceName(r.Connection),
			Name:       r.Connection,
			Attributes: attributes,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	if r.Datasource {
		tmpl, err := template.New("datasource.tf.go.tmpl").ParseFiles(exampleDatasourceTemplate)
		if err != nil {
			return err
		}
		newpath := filepath.Join(
			exampleDatasourceOutputPath,
			fmt.Sprintf("polytomic_%s_connection", r.Connection),
		)
		err = os.MkdirAll(newpath, os.ModePerm)
		if err != nil {
			return err
		}
		f, err := os.Create(
			filepath.Join(newpath, "data-source.tf"))

		if err != nil {
			return err
		}
		defer f.Close()

		err = tmpl.Execute(f, struct {
			Resource   string
			Name       string
			Attributes []Attribute
		}{
			Resource:   TerraformResourceName(r.Connection),
			Name:       r.Connection,
			Attributes: attributes,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func writeConnectionResource(r Connection) error {
	tmpl, err := template.New("resource.go.tmpl").ParseFiles(connectionResourceTemplate)
	if err != nil {
		log.Fatal(fmt.Errorf("error parsing resource template: %w", err))
	}
	var buf bytes.Buffer
	f, err := os.Create(
		filepath.Join(outputPath, fmt.Sprintf("resource_%s_connection.go", r.Connection)),
	)
	if err != nil {
		log.Fatal(err)
	}

	imports := defaultImports
	for k := range r.ExtraImports {
		imports += fmt.Sprintf("\n\"%s\"", k)
	}

	defer f.Close()
	err = tmpl.Execute(&buf, Connection{
		Name:         r.Name,
		Conn:         r.Connection,
		Connection:   strings.Title(r.Connection),
		ResourceName: r.Connection,
		Attributes:   r.Attributes,
		Type:         r.Type,
		Config:       r.Config,
		Imports:      imports,
	})
	if err != nil {
		log.Fatal(fmt.Errorf("error executing resource template: %w", err))
	}
	_, err = f.Write(buf.Bytes())

	p, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(fmt.Errorf("error formatting resource %s: %w", r.Connection, err))
	}
	f.Close()
	f, err = os.Create(f.Name())

	_, err = f.Write(p)
	return err
}

func writeConnectionDataSource(r Connection) error {
	tmpl, err := template.New("datasource.go.tmpl").ParseFiles(connectionDataSourceTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	f, err := os.Create(
		filepath.Join(outputPath, fmt.Sprintf("datasource_%s_connection.go", r.Connection)))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var attributes []Attribute
	for _, a := range r.Attributes {
		if !a.Sensitive {
			attributes = append(attributes, a)
		}
	}

	err = tmpl.Execute(&buf, Connection{
		Name:         r.Name,
		Connection:   strings.Title(r.Connection),
		ResourceName: r.Connection,
		Attributes:   attributes,
		Type:         r.Type,
		Config:       r.Config,
	})
	if err != nil {
		log.Fatal(err)
	}
	p, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.Write(p)
	return err
}

func writeExports(datasources, resources []Importable) error {
	slices.SortFunc(datasources, func(a, b Importable) int {
		return cmp.Compare(a.Name, b.Name)
	})
	slices.SortFunc(resources, func(a, b Importable) int {
		return cmp.Compare(a.Name, b.Name)
	})

	tmpl, err := template.New("connections.go.tmpl").ParseFiles(exportTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	f, err := os.Create(
		filepath.Join(outputPath, "connections.go"))
	if err != nil {
		return err
	}
	defer f.Close()
	err = tmpl.Execute(&buf, struct {
		Datasources []Importable
		Resources   []Importable
	}{
		Datasources: datasources,
		Resources:   resources,
	})
	if err != nil {
		log.Fatal(err)
	}
	p, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.Write(p)
	return err
}

func TerraformResourceName(connection string) string {
	return fmt.Sprintf("polytomic_%s_connection", connection)
}

const (
	legalCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
)

// A name must start with a letter or underscore and
// may contain only letters, digits, underscores, and dashes.
// e.g 100_users -> _100_users
func ValidName(s string) string {
	if len(s) == 0 {
		return "_"
	}

	// if string is not a letter or underscore, prepend underscore
	if !unicode.IsLetter(rune(s[0])) && s[0] != '_' {
		s = "_" + s
	}

	// replace illegal characters with underscore
	for i, v := range []byte(s) {
		if !strings.Contains(legalCharacters, string(v)) {
			s = s[:i] + "_" + s[i+1:]
		}
		if unicode.IsLower(rune(v)) && i < len(s)-1 && unicode.IsUpper(rune(s[i+1])) {
			s = s[:i+1] + "_" + strings.ToLower(s[i+1:])
		}
	}

	return s
}
