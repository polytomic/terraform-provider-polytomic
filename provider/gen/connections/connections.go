package connections

import (
	"bytes"
	"cmp"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

const (
	// General
	ConnectionsFile = "./provider/gen/connections/connections.yaml"
	outputPath      = "./provider"
	exportTemplate  = "./provider/gen/connections/connections.go.tmpl"

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
		"string": {
			AttrType:      "schema.StringAttribute",
			TfType:        "types.String",
			NewAttrType:   "types.StringType",
			Default:       "stringdefault.StaticString(\"\")",
			DefaultImport: "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault",
		},
		"number": {
			AttrType:      "schema.NumberAttribute",
			TfType:        "types.Number",
			NewAttrType:   "types.NumberType",
			Default:       "int64default.StaticInt64(0)",
			DefaultImport: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
		},
		"bool": {
			AttrType:    "schema.BoolAttribute",
			TfType:      "types.Bool",
			NewAttrType: "types.BoolType",
		},
		"int": {
			AttrType:      "schema.Int64Attribute",
			TfType:        "types.Int64",
			NewAttrType:   "types.NumberType",
			Default:       "int64default.StaticInt64(0)",
			DefaultImport: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
		},
		"int64": {
			AttrType:      "schema.Int64Attribute",
			TfType:        "types.Int64",
			NewAttrType:   "types.NumberType",
			Default:       "int64default.StaticInt64(0)",
			DefaultImport: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
		},
	}
)

type Typer struct {
	AttrType      string
	TfType        string
	NewAttrType   string
	Default       string
	DefaultImport string
}

type Connections struct {
	Connections []Connection `yaml:"connections"`
}

type Connection struct {
	Connection   string          `yaml:"connection"`
	Name         string          `yaml:"name"`
	Conn         string          `yaml:"-"`
	ResourceName string          `yaml:"-"`
	Type         string          `yaml:"type,omitempty"`
	Attributes   []Attribute     `yaml:"attributes"`
	Datasource   bool            `yaml:"datasource,omitempty"`
	Resource     bool            `yaml:"resource,omitempty"`
	ExtraImports map[string]bool `yaml:"-"`
	Imports      string          `yaml:"-"`
}

type Attribute struct {
	Name                string `yaml:"name"`
	CapName             string `yaml:"-"`
	NameOverride        string `yaml:"name_override,omitempty"`
	Alias               string `yaml:"alias,omitempty"`
	Sensitive           bool   `yaml:"sensitive,omitempty"`
	Required            bool   `yaml:"required,omitempty"`
	Optional            bool   `yaml:"optional,omitempty"`
	Computed            bool   `yaml:"computed,omitempty"`
	Type                string `yaml:"type"`
	Description         string `yaml:"description,omitempty"`
	Example             string `yaml:"example,omitempty"`
	ExampleTypeOverride string `yaml:"example_type_override,omitempty"`

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

"github.com/AlekSi/pointer"
"github.com/hashicorp/terraform-plugin-framework/attr"
"github.com/hashicorp/terraform-plugin-framework/path"
"github.com/hashicorp/terraform-plugin-framework/resource"
"github.com/hashicorp/terraform-plugin-framework/resource/schema"
"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
"github.com/hashicorp/terraform-plugin-framework/types"
"github.com/hashicorp/terraform-plugin-log/tflog"
"github.com/polytomic/polytomic-go"
"github.com/mitchellh/mapstructure"
ptclient "github.com/polytomic/polytomic-go/client"
ptcore "github.com/polytomic/polytomic-go/core"
`

type Importable struct {
	Name         string
	ResourceName string
}

func SortConnections() error {
	config, err := os.ReadFile(ConnectionsFile)
	if err != nil {
		return err
	}
	data := Connections{}
	err = yaml.Unmarshal(config, &data)
	if err != nil {
		return err
	}
	slices.SortFunc(data.Connections, func(a, b Connection) int {
		return cmp.Compare(a.Connection, b.Connection)
	})

	result, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(ConnectionsFile, result, 0644)
}

func GenerateConnections() error {
	config, err := os.ReadFile(ConnectionsFile)
	if err != nil {
		return err
	}
	data := Connections{}
	err = yaml.Unmarshal(config, &data)
	if err != nil {
		return err
	}

	resources := []Importable{}
	datasources := []Importable{}

	for _, r := range data.Connections {
		r.ExtraImports = make(map[string]bool)
		for i, a := range r.Attributes {
			t, ok := TypeMap[a.Type]
			if !ok {
				return fmt.Errorf("type %s not found for %s", a.Type, r.Name)
			}
			r.Attributes[i].TfType = t.TfType
			r.Attributes[i].AttrType = t.AttrType
			r.Attributes[i].NewAttrType = t.NewAttrType
			r.Attributes[i].AttrName = a.Name
			r.Attributes[i].CapName = strings.Title(a.Name)
			if a.NameOverride != "" {
				r.Attributes[i].AttrName = a.NameOverride
			}
			r.Attributes[i].Computed = a.Computed || a.Optional
			if !a.Required {
				r.Attributes[i].Default = t.Default
				if t.DefaultImport != "" {
					r.ExtraImports[t.DefaultImport] = true
				}
			}
		}
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
	for i, a := range r.Attributes {
		if a.ExampleTypeOverride != "" {
			r.Attributes[i].Type = a.ExampleTypeOverride
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
			filepath.Join(newpath, fmt.Sprintf("resource.tf")))

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
		log.Fatal(err)
	}
	var buf bytes.Buffer
	f, err := os.Create(
		filepath.Join(outputPath, fmt.Sprintf("resource_%s_connection.go", r.Connection)))
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
		Imports:      imports,
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

func writeConnectionDataSource(r Connection) error {
	tmpl, err := template.New("datasource.go.tmpl").ParseFiles(connectionDataSourceTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	f, err := os.Create(
		filepath.Join(outputPath, fmt.Sprintf("datasource_%s_connection.go", r.Connection)))
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
