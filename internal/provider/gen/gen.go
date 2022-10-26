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
	resourceTemplate  = "./internal/provider/gen/connections/connection.go.tmpl"
	exportTemplate    = "./internal/provider/gen/connections/resources.go.tmpl"
	exampleTemplate   = "./internal/provider/gen/connections/resource.tf.go.tmpl"
	connectionsFile   = "./internal/provider/gen/connections.yaml"
	outputPath        = "./internal/provider"
	exampleOutputPath = "./examples/resources"
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
}

type attribute struct {
	Name        string `yaml:"name"`
	Sensitive   bool   `yaml:"sensitive"`
	Required    bool   `yaml:"required"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Example     string `yaml:"example"`

	TfType   string `yaml:"-"`
	AttrType string `yaml:"-"`
	AttrName string `yaml:"-"`
}

type export struct {
	Name string
	Type string
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
	exports := []export{}
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
		err = writeConnectionResource(r)
		if err != nil {
			log.Fatal(err)
		}
		err = writeConnectionExample(r)
		if err != nil {
			log.Fatal(err)
		}
		exports = append(exports, export{
			Name: terraformResourceName(r.Connection),
			Type: fmt.Sprintf("%sConnectionResourceType{}", r.Connection),
		})
	}
	err = writeResourceExports(exports)
	if err != nil {
		log.Fatal(err)
	}
}

func writeConnectionExample(r connection) error {
	tmpl, err := template.New("resource.tf.go.tmpl").ParseFiles(exampleTemplate)
	if err != nil {
		log.Fatal(err)
	}
	newpath := filepath.Join(
		exampleOutputPath,
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
		Attributes []attribute
	}{
		Resource:   terraformResourceName(r.Connection),
		Name:       r.Connection,
		Attributes: r.Attributes,
	})
	if err != nil {
		log.Fatal(err)
	}

	return err
}

func writeConnectionResource(r connection) error {
	tmpl, err := template.New("connection.go.tmpl").ParseFiles(resourceTemplate)
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

func writeResourceExports(exports []export) error {
	tmpl, err := template.New("resources.go.tmpl").ParseFiles(exportTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	f, err := os.Create(filepath.Join(outputPath, "connection_resources.go"))
	defer f.Close()
	err = tmpl.Execute(&buf, struct {
		Resources []export
	}{
		Resources: exports,
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
