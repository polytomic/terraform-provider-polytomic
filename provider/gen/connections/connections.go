package connections

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/polytomic/terraform-provider-polytomic/provider"
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
			AttrType:    "schema.StringAttribute",
			TfType:      "types.String",
			NewAttrType: "types.StringType",
			Default:     "stringdefault.StaticString(\"\")",
		},
		"number": {
			AttrType:    "schema.NumberAttribute",
			TfType:      "types.Number",
			NewAttrType: "types.NumberType",
			Default:     "int64default.StaticInt64(0)",
		},
		"bool": {
			AttrType:    "schema.BoolAttribute",
			TfType:      "types.Bool",
			NewAttrType: "types.BoolType",
		},
		"int": {
			AttrType:    "schema.Int64Attribute",
			TfType:      "types.Int64",
			NewAttrType: "types.NumberType",
			Default:     "int64default.StaticInt64(0)",
		},
		"int64": {
			AttrType:    "schema.Int64Attribute",
			TfType:      "types.Int64",
			NewAttrType: "types.NumberType",
			Default:     "int64default.StaticInt64(0)",
		},
	}
)

type Typer struct {
	AttrType    string
	TfType      string
	NewAttrType string
	Default     string
}

type Connections struct {
	Connections []Connection `yaml:"connections"`
}

type Connection struct {
	Name         string `yaml:"name"`
	Connection   string `yaml:"connection"`
	ResourceName string
	Type         string      `yaml:"type"`
	Attributes   []Attribute `yaml:"attributes"`
	Config       string      `yaml:"config"`
	Datasource   bool        `yaml:"datasource"`
	Resource     bool        `yaml:"resource"`
}

type Attribute struct {
	Name                string `yaml:"name"`
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

type Importable struct {
	Name         string
	ResourceName string
}

func GenerateConnections() error {
	config, err := ioutil.ReadFile(ConnectionsFile)
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
		for i, a := range r.Attributes {
			t, ok := TypeMap[a.Type]
			if !ok {
				return fmt.Errorf("type %s not found for %s", a.Type, r.Name)
			}
			r.Attributes[i].TfType = t.TfType
			r.Attributes[i].AttrType = t.AttrType
			r.Attributes[i].NewAttrType = t.NewAttrType
			r.Attributes[i].AttrName = provider.ToSnakeCase(a.Name)
			if a.NameOverride != "" {
				r.Attributes[i].AttrName = a.NameOverride
			}
			r.Attributes[i].Computed = a.Computed || a.Optional
			if !a.Required {
				r.Attributes[i].Default = t.Default
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
			resources = append(resources, Importable{
				Name:         r.Connection,
				ResourceName: fmt.Sprintf("%sConnectionResource", strings.Title(r.Connection)),
			})
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
			filepath.Join(newpath, fmt.Sprintf("data-source.tf")))

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
	defer f.Close()
	err = tmpl.Execute(&buf, Connection{
		Name:         r.Name,
		Connection:   strings.Title(r.Connection),
		ResourceName: r.Connection,
		Attributes:   r.Attributes,
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
