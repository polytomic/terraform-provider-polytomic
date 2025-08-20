package connections

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"log"
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
	outputPath      = "./provider/internal/connections"
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
			AttrType:     "schema.SetAttribute",
			TfType:       "Set",
			ReadAttrType: "types.SetType",
			GoType:       "[]",
		},
		"object": {
			AttrType:     "schema.SingleNestedAttribute",
			TfType:       "Object",
			ReadAttrType: "types.ObjectType",
			GoType:       "struct",
		},
		"": {
			AttrType:     "schema.StringAttribute",
			TfType:       "String",
			ReadAttrType: "types.StringType",

			GoType: "string",
		},
		"string": {
			AttrType:     "schema.StringAttribute",
			TfType:       "String",
			ReadAttrType: "types.StringType",
			Default: DefaultValue{
				Value:  "stringdefault.StaticString(\"\")",
				Import: "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault",
			},
			GoType: "string",
		},
		"number": {
			AttrType:     "schema.NumberAttribute",
			TfType:       "Number",
			ReadAttrType: "types.NumberType",
			Default: DefaultValue{
				Value:  "int64default.StaticInt64(0)",
				Import: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
			},
			GoType: "int64",
		},
		"bool": {
			AttrType:     "schema.BoolAttribute",
			TfType:       "Bool",
			ReadAttrType: "types.BoolType",
			GoType:       "bool",
		},
		"boolean": {
			AttrType:     "schema.BoolAttribute",
			TfType:       "Bool",
			ReadAttrType: "types.BoolType",
			GoType:       "bool",
		},
		"int": {
			AttrType:     "schema.Int64Attribute",
			TfType:       "Int64",
			ReadAttrType: "types.NumberType",
			Default: DefaultValue{
				Value:  "int64default.StaticInt64(0)",
				Import: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
			},
			GoType: "int64",
		},
		"int64": {
			AttrType:     "schema.Int64Attribute",
			TfType:       "Int64",
			ReadAttrType: "types.NumberType",
			Default: DefaultValue{
				Value:  "int64default.StaticInt64(0)",
				Import: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
			},
			GoType: "int64",
		},
		"integer": {
			AttrType:     "schema.Int64Attribute",
			TfType:       "Int64",
			ReadAttrType: "types.NumberType",
			Default: DefaultValue{
				Value:  "int64default.StaticInt64(0)",
				Import: "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default",
			},
			GoType: "int64",
		},
	}
)

type DefaultValue struct {
	Value  string
	Import string
}

type Typer struct {
	AttrType     string
	TfType       string
	ReadAttrType string
	Default      DefaultValue
	GoType       string
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
	Sensitive           bool   `yaml:"sensitive"`
	Required            bool   `yaml:"required"`
	Optional            bool   `yaml:"optional"`
	Computed            bool   `yaml:"computed"`
	Type                string `yaml:"type"`
	Description         string `yaml:"description"`
	Example             string `yaml:"example"`
	ExampleTypeOverride string `yaml:"example_type_override"`

	TfType string `yaml:"-"`
	// AttrType is the Terraform schema.* type for the attribute.
	AttrType     string       `yaml:"-"`
	AttrReadType string       `yaml:"-"`
	AttrName     string       `yaml:"-"`
	Default      DefaultValue `yaml:"-"`
	Attributes   []Attribute
	Elem         *Attribute
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
	client := getPTClient()
	data, err := fetchOrRead(ctx,
		connectionTypes,
		func(ctx context.Context) (*polytomic.ConnectionTypeResponseEnvelope, error) {
			return client.Connections.GetTypes(ctx)
		},
	)
	if err != nil {
		return err
	}

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
		js, err := getJSONSchema(connSchema)
		if err != nil {
			return fmt.Errorf("error converting API response to jsonschema: %w", err)
		}
		attrs, err := attributesForJSONSchema(js)
		if err != nil {
			return fmt.Errorf("error inspecting attributes for %s: %w", r.Connection, err)
		}
		for _, a := range attrs {
			if a.Default.Import != "" {
				r.ExtraImports[a.Default.Import] = true
			}
		}
		r.Attributes = append(r.Attributes, attrs...)
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

func getPTClient() *ptclient.Client {
	baseURL, ok := os.LookupEnv("POLYTOMIC_DEPLOYMENT_URL")
	if !ok {
		fmt.Println("POLYTOMIC_DEPLOYMENT_URL not set; using production.")
	}
	apiKey, ok := os.LookupEnv("POLYTOMIC_API_KEY")
	if !ok {
		fmt.Println("POLYTOMIC_API_KEY not set; using cached connection definitions.")
	}
	client := ptclient.NewClient(
		option.WithBaseURL(baseURL),
		option.WithToken(apiKey),
	)
	return client
}

func attributesForJSONSchema(connSchema *jsonschema.Schema) ([]Attribute, error) {
	attrs := []Attribute{}
	for pair := connSchema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		attr, err := tfAttr(pair.Key, pair.Value, connSchema.Required)
		if err != nil {
			return attrs, err
		}

		attrs = append(attrs, attr)
	}
	return attrs, nil
}

func tfAttr(k string, a *jsonschema.Schema, required []string) (Attribute, error) {
	t, ok := TypeMap[a.Type]
	if !ok {
		return Attribute{}, fmt.Errorf("type %s not found for %s", a.Type, k)
	}
	var ex string
	if len(a.Examples) > 0 {
		if exstr, ok := a.Examples[0].(string); ok {
			ex = exstr
		}
	}

	title := strings.TrimSpace(strings.TrimSuffix(a.Title, "(optional)"))
	desc := ""
	if !strings.EqualFold(title, ValidName(k)) {
		desc = title
		desc += "\n\n"
	}
	desc += fmt.Sprintf("    %s", a.Description)
	desc = strings.TrimSpace(desc)
	attr := Attribute{
		TfType:       t.TfType,
		AttrType:     t.AttrType,
		AttrReadType: t.ReadAttrType,
		AttrName:     ValidName(k), // key in the tf schema
		CapName:      strings.Title(k),
		Name:         k, // key in the payload
		Type:         t.GoType,
		Description:  desc,
		Example:      ex,
		Sensitive:    a.Extras["sensitive"] == true,
	}
	if a.Format == "json" && attr.Example != "" {
		attr.Example = fmt.Sprintf("jsonencode(%s)", attr.Example)
		attr.ExampleTypeOverride = "json"
	}
	switch a.Type {
	case "array":
		elem, err := tfAttr(k, a.Items, a.Items.Required)
		if err != nil {
			return Attribute{}, fmt.Errorf("error inspecting attributes for %s: %w", k, err)
		}
		switch a.Items.Type {
		case "object":
			attr.AttrType = "schema.SetNestedAttribute"
		default:
			attr.AttrReadType = "types.SetType"
		}
		attr.Elem = &elem
	case "object":
		sa, err := attributesForJSONSchema(a)
		if err != nil {
			return Attribute{}, fmt.Errorf("error inspecting attributes for %s: %w", k, err)
		}
		attr.Attributes = sa
	}
	attr.Required = slices.Contains(required, k)
	attr.Optional = !attr.Required
	attr.Computed = a.ReadOnly || attr.Optional
	if attr.Computed {
		attr.Default = t.Default
	}
	return attr, nil
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

func valueAttr(a Attribute) string {
	b := &strings.Builder{}
	va(nil, a, b)

	return b.String()
}

func va(prefix []string, a Attribute, builder *strings.Builder) {
	// "password": data.Configuration.Attributes()["auth"].(types.Object).Attributes()["basic"].(types.Object).Attributes()["password"].(types.String).ValueString(),
	builder.WriteString(fmt.Sprintf("\"%s\": ", a.Name))
	switch a.Type {
	case "int", "integer", "int64", "number":
		builder.WriteString("int(")
	case "map[string]interface{}":
		fmt.Fprintln(builder, "map[string]interface{}{")
		ap := append([]string{}, prefix...)
		ap = append(ap, a.AttrName)
		for _, aa := range a.Attributes {
			va(ap, aa, builder)
		}
		fmt.Fprintln(builder, "},")
		return
	}
	builder.WriteString("data.Configuration.Attributes()")
	for _, p := range prefix {
		fmt.Fprintf(builder, `["%s"].(types.Object).Attributes()`, p)
	}
	fmt.Fprintf(builder, "[\"%s\"]", a.Name)

	switch a.Type {
	case "int", "integer", "int64", "number":
		fmt.Fprintf(builder, ".(types.%s).ValueInt64()),\n", a.TfType)
	case "bool":
		fmt.Fprintf(builder, ".(types.%s).ValueBool(),\n", a.TfType)
	case "string":
		fmt.Fprintf(builder, ".(types.%s).ValueString(),\n", a.TfType)

	}
}

func writeConnectionResource(r Connection) error {
	tmpl, err := template.New("resource.go.tmpl").
		Funcs(template.FuncMap{
			"valueAttr": valueAttr,
			"lower":     strings.ToLower,
		}).
		ParseFiles(connectionResourceTemplate)
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
	if err != nil {
		log.Fatal(fmt.Errorf("error writing resource %s: %w", r.Connection, err))
	}

	p, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(fmt.Errorf("error formatting resource %s: %w", r.Connection, err))
	}
	f.Close()
	f, err = os.Create(f.Name())
	if err != nil {
		log.Fatal(fmt.Errorf("error creating resource %s: %w", r.Connection, err))
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
