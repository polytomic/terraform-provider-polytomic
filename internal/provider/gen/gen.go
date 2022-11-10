package main

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

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/terraform-provider-polytomic/internal/provider"
	"gopkg.in/yaml.v2"
)

const (
	// General
	connectionsFile = "./internal/provider/gen/connections.yaml"
	outputPath      = "./internal/provider"
	exportTemplate  = "./internal/provider/gen/connections/connections.go.tmpl"

	// Resources
	connectionResourceTemplate = "./internal/provider/gen/connections/resource.go.tmpl"
	exampleResourceTemplate    = "./internal/provider/gen/connections/resource.tf.go.tmpl"
	exampleResourceOutputPath  = "./examples/resources"

	// Datasources
	connectionDataSourceTemplate = "./internal/provider/gen/connections/datasource.go.tmpl"
	exampleDatasourceTemplate    = "./internal/provider/gen/connections/datasource.tf.go.tmpl"
	exampleDatasourceOutputPath  = "./examples/data-sources"
)

var (
	typeMap = map[string]typer{
		"string": {
			AttrType: types.StringType.String(),
			TfType:   "types.String",
		},
		"number": {
			AttrType: types.NumberType.String(),
			TfType:   "types.Number",
		},
		"bool": {
			AttrType: types.BoolType.String(),
			TfType:   "types.Bool",
		},
		"int": {
			AttrType: types.Int64Type.String(),
			TfType:   "types.Int64",
		},
		"int64": {
			AttrType: types.Int64Type.String(),
			TfType:   "types.Int64",
		},
	}
)

type typer struct {
	AttrType string
	TfType   string
}

type connections struct {
	Connections []connection `yaml:"connections"`
}

type connection struct {
	Name       string      `yaml:"name"`
	Connection string      `yaml:"connection"`
	Type       string      `yaml:"type"`
	Attributes []attribute `yaml:"attributes"`
	Config     string      `yaml:"config"`
	Datasource bool        `yaml:"datasource"`
	Resource   bool        `yaml:"resource"`
}

type attribute struct {
	Name                string `yaml:"name"`
	Sensitive           bool   `yaml:"sensitive"`
	Required            bool   `yaml:"required"`
	Optional            bool   `yaml:"optional"`
	Type                string `yaml:"type"`
	Description         string `yaml:"description"`
	Example             string `yaml:"example"`
	ExampleTypeOverride string `yaml:"example_type_override"`

	TfType   string `yaml:"-"`
	AttrType string `yaml:"-"`
	AttrName string `yaml:"-"`
}

func main() {
	config, err := ioutil.ReadFile(connectionsFile)
	if err != nil {
		log.Fatal(err)
	}
	data := connections{}
	err = yaml.Unmarshal(config, &data)
	if err != nil {
		log.Fatal(err)
	}
	resources := []string{}
	datasources := []string{}
	for _, r := range data.Connections {
		for i, a := range r.Attributes {
			t, ok := typeMap[a.Type]
			if !ok {
				log.Fatalf("unknown type %s", a.Type)
			}
			r.Attributes[i].TfType = t.TfType
			r.Attributes[i].AttrType = t.AttrType
			r.Attributes[i].AttrName = provider.ToSnakeCase(a.Name)
		}
		if r.Name == "" {
			r.Name = strings.Title(r.Connection)
		}
		if r.Resource {
			err := writeConnectionResource(r)
			if err != nil {
				log.Fatal(err)
			}
			resources = append(resources, fmt.Sprintf("%sConnectionResource", r.Connection))
		}
		if r.Datasource {
			err := writeConnectionDataSource(r)
			if err != nil {
				log.Fatal(err)
			}
			datasources = append(datasources, fmt.Sprintf("%sConnectionDataSource", r.Connection))
		}

		err = writeConnectionExamples(r)
		if err != nil {
			log.Fatal(err)
		}

	}

	err = writeExports(datasources, resources)
	if err != nil {
		log.Fatal(err)
	}

}

func writeConnectionExamples(r connection) error {

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

		// Overrides types for examples
		for i, a := range r.Attributes {
			if a.ExampleTypeOverride != "" {
				r.Attributes[i].Type = a.ExampleTypeOverride
			}
		}

		err = tmpl.Execute(f, struct {
			Resource   string
			Name       string
			Attributes []attribute
		}{
			Resource:   terraformResourceName(r.Connection),
			Name:       r.Connection,
			Attributes: r.Attributes,
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

		// Overrides types for examples
		for i, a := range r.Attributes {
			if a.ExampleTypeOverride != "" {
				r.Attributes[i].Type = a.ExampleTypeOverride
			}
		}

		err = tmpl.Execute(f, struct {
			Resource   string
			Name       string
			Attributes []attribute
		}{
			Resource:   terraformResourceName(r.Connection),
			Name:       r.Connection,
			Attributes: r.Attributes,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func writeConnectionResource(r connection) error {
	tmpl, err := template.New("resource.go.tmpl").ParseFiles(connectionResourceTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	f, err := os.Create(
		filepath.Join(outputPath, fmt.Sprintf("resource_%s_connection.go", r.Connection)))
	defer f.Close()
	err = tmpl.Execute(&buf, connection{
		Name:       r.Name,
		Connection: r.Connection,
		Attributes: r.Attributes,
		Type:       r.Type,
		Config:     r.Config,
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

func writeConnectionDataSource(r connection) error {
	tmpl, err := template.New("datasource.go.tmpl").ParseFiles(connectionDataSourceTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	f, err := os.Create(
		filepath.Join(outputPath, fmt.Sprintf("datasource_%s_connection.go", r.Connection)))
	defer f.Close()
	err = tmpl.Execute(&buf, connection{
		Name:       r.Name,
		Connection: r.Connection,
		Attributes: r.Attributes,
		Type:       r.Type,
		Config:     r.Config,
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

func writeExports(datasources, resources []string) error {
	tmpl, err := template.New("connections.go.tmpl").ParseFiles(exportTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	f, err := os.Create(
		filepath.Join(outputPath, "connections.go"))
	defer f.Close()
	err = tmpl.Execute(&buf, struct {
		Datasources []string
		Resources   []string
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

func terraformResourceName(connection string) string {
	return fmt.Sprintf("polytomic_%s_connection", connection)
}
