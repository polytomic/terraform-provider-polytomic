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
	connectionResourceTemplate    = "./provider/gen/connections/resource.go.tmpl"
	connectionResourceDocTemplate = "./provider/gen/connections/resource_doc.md.tmpl"
	exampleResourceTemplate       = "./provider/gen/connections/resource.tf.go.tmpl"
	exampleResourceOutputPath     = "./examples/resources"
	docTemplateOutputPath         = "./templates/resources"
	forceDestroyDescriptionFile   = "./provider/internal/connections/force_destroy.md"

	// Datasources
	connectionDataSourceTemplate = "./provider/gen/connections/datasource.go.tmpl"
	exampleDatasourceTemplate    = "./provider/gen/connections/datasource.tf.go.tmpl"
	exampleDatasourceOutputPath  = "./examples/data-sources"
)

// blocklist contains backend IDs that should not have resources or data sources generated
var blocklist = map[string]bool{
	// Add backend IDs to exclude here
	"fakedata":    true,
	"localsqlite": true,
}

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
		"map": {
			AttrType:     "schema.MapAttribute",
			TfType:       "Map",
			ReadAttrType: "types.MapType",
			GoType:       "map[string]string",
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

// AttrCondition describes when an attribute is applicable, based on
// the value of another field in the same configuration block.
type AttrCondition struct {
	// Field is the name of the field this condition depends on.
	Field string
	// Value is the value that Field must have. Nil means "any value" (presence check).
	Value interface{}
	// Required indicates the attribute is required (not just visible) under this condition.
	Required bool
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
	AttrType     string          `yaml:"-"`
	AttrReadType string          `yaml:"-"`
	AttrName     string          `yaml:"-"`
	Default      DefaultValue    `yaml:"-"`
	EnumValues   []string        `yaml:"-"` // valid values for string enums
	EnumLabels   []string        `yaml:"-"` // human-readable labels for enum values (parallel to EnumValues)
	Conditions   []AttrCondition `yaml:"-"` // conditions under which this attribute applies
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
"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
"github.com/hashicorp/terraform-plugin-framework/types"
"github.com/hashicorp/terraform-plugin-log/tflog"
"github.com/polytomic/polytomic-go"
ptcore "github.com/polytomic/polytomic-go/core"
"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
`

type Importable struct {
	Name         string
	ResourceName string
}

type ConnectionType struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	UseOAuth bool   `json:"use_oauth"`
}

var useCache = os.Getenv("POLYTOMIC_USE_CACHE") != ""

func readCached[T any](path string) (T, error) {
	var data T
	ct, err := os.Open(path)
	if err != nil {
		return data, fmt.Errorf("error opening cached %s: %w", path, err)
	}
	defer ct.Close()
	if err := json.NewDecoder(ct).Decode(&data); err != nil {
		return data, fmt.Errorf("error reading %s: %w", path, err)
	}
	return data, nil
}

func fetchOrRead[T any](ctx context.Context, path string, fetch func(context.Context) (T, error)) (T, error) {
	if useCache {
		return readCached[T](path)
	}

	data, err := fetch(ctx)
	if err != nil {
		// an error occurred fetching; see if we have a cached copy
		return readCached[T](path)
	}

	// write the fetched data to path
	f, err := os.Create(path)
	if err != nil {
		return *(new(T)), fmt.Errorf("error creating %s: %w", path, err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(data)
	if err != nil {
		return *(new(T)), fmt.Errorf("error encoding %s: %w", path, err)
	}
	return data, nil
}

func GenerateConnections(ctx context.Context) error {
	client := getPTClient()
	data, err := fetchOrRead(ctx,
		connectionTypes,
		func(ctx context.Context) ([]ConnectionType, error) {
			connTypes, err := client.Connections.GetTypes(ctx)
			if err != nil {
				return nil, err
			}

			result := make([]ConnectionType, len(connTypes.Data))
			for i, ct := range connTypes.Data {
				result[i] = ConnectionType{
					ID:       pointer.Get(ct.Id),
					Name:     pointer.Get(ct.Name),
					UseOAuth: pointer.Get(ct.UseOauth),
				}
			}

			return result, nil
		},
	)
	if err != nil {
		return err
	}

	resources := []Importable{}
	datasources := []Importable{}

	for _, connType := range data {
		// Skip blocklisted connection types
		if blocklist[connType.ID] {
			log.Printf("Skipping blocklisted connection type: %s", connType.ID)
			continue
		}

		connSchema, err := fetchOrRead(ctx,
			filepath.Join(jsonschemaPath, fmt.Sprintf("%s.json", connType.ID)),
			func(ctx context.Context) (*polytomic.JsonschemaSchema, error) {
				return client.Connections.GetConnectionTypeSchema(ctx, connType.ID)
			},
		)
		if err != nil {
			log.Printf("Skipping connection type %s: %v", connType.ID, err)
			continue
		}
		r := Connection{
			Name:         cmp.Or(connType.Name, connType.ID),
			ResourceName: connType.ID,
			Connection:   connType.ID,
			Type:         connType.ID,
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
			if len(a.EnumValues) > 0 {
				r.ExtraImports["github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"] = true
				r.ExtraImports["github.com/hashicorp/terraform-plugin-framework/schema/validator"] = true
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

		if r.Resource {
			err = writeConnectionDocTemplate(r)
			if err != nil {
				return err
			}
		}

	}

	err = writeExports(datasources, resources)
	if err != nil {
		return err
	}

	// Build the set of connection IDs that were generated so we can
	// remove orphaned artifacts from previous runs.
	generated := make(map[string]bool, len(data))
	for _, ct := range data {
		if !blocklist[ct.ID] {
			generated[ct.ID] = true
		}
	}
	if err := cleanupOrphanedConnections(generated); err != nil {
		return fmt.Errorf("error cleaning up orphaned connections: %w", err)
	}

	return nil
}

// cleanupOrphanedConnections removes generated files for connection types
// that no longer exist in the API response.
func cleanupOrphanedConnections(generated map[string]bool) error {
	// Each entry maps a directory to a pattern that extracts the connection ID
	// from the filename. We only touch files that match the connection naming
	// convention so hand-written files are never deleted.
	type cleanupTarget struct {
		dir    string
		prefix string // filename prefix before the connection ID
		suffix string // filename suffix after the connection ID
		isDir  bool   // true if the artifact is a directory, not a file
	}

	targets := []cleanupTarget{
		{dir: outputPath, prefix: "resource_", suffix: "_connection.go"},
		{dir: outputPath, prefix: "datasource_", suffix: "_connection.go"},
		{dir: docTemplateOutputPath, prefix: "", suffix: "_connection.md.tmpl"},
		{dir: jsonschemaPath, prefix: "", suffix: ".json"},
		{dir: exampleResourceOutputPath, prefix: "polytomic_", suffix: "_connection", isDir: true},
		{dir: exampleDatasourceOutputPath, prefix: "polytomic_", suffix: "_connection", isDir: true},
	}

	for _, t := range targets {
		entries, err := os.ReadDir(t.dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		for _, e := range entries {
			name := e.Name()
			if !strings.HasPrefix(name, t.prefix) || !strings.HasSuffix(name, t.suffix) {
				continue
			}
			connID := strings.TrimPrefix(name, t.prefix)
			connID = strings.TrimSuffix(connID, t.suffix)
			if connID == "" {
				continue
			}
			if generated[connID] {
				continue
			}
			path := filepath.Join(t.dir, name)
			log.Printf("Removing orphaned artifact: %s", path)
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("error removing %s: %w", path, err)
			}
		}
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
	// Track which attribute names we've already seen (by index) so
	// dependentSchemas doesn't introduce duplicates and conditions merge.
	seen := map[string]int{}
	for pair := connSchema.Properties.Oldest(); pair != nil; pair = pair.Next() {
		attr, err := tfAttr(pair.Key, pair.Value, connSchema.Required)
		if err != nil {
			return attrs, err
		}
		seen[pair.Key] = len(attrs)
		attrs = append(attrs, attr)
	}

	// Extract conditional attributes from dependentSchemas.
	// The same attribute may appear under multiple conditions (e.g., a field
	// shown for two different enum values), so we merge conditions.
	depAttrs, err := attributesFromDependentSchemas(connSchema)
	if err != nil {
		return attrs, err
	}
	for _, attr := range depAttrs {
		if idx, ok := seen[attr.Name]; ok {
			// Merge conditions into the existing attribute.
			attrs[idx].Conditions = append(attrs[idx].Conditions, attr.Conditions...)
		} else {
			seen[attr.Name] = len(attrs)
			attrs = append(attrs, attr)
		}
	}

	// Annotate descriptions with condition info now that all conditions
	// have been merged.
	for i := range attrs {
		annotateConditionDescription(&attrs[i])
	}

	return attrs, nil
}

// annotateConditionDescription appends a human-readable note to the attribute
// description when the attribute has conditions attached.
func annotateConditionDescription(attr *Attribute) {
	if len(attr.Conditions) == 0 {
		return
	}
	// Group conditions by field for cleaner output.
	byField := map[string][]AttrCondition{}
	var fieldOrder []string
	for _, c := range attr.Conditions {
		if _, ok := byField[c.Field]; !ok {
			fieldOrder = append(fieldOrder, c.Field)
		}
		byField[c.Field] = append(byField[c.Field], c)
	}

	var parts []string
	for _, field := range fieldOrder {
		conds := byField[field]
		values := []string{}
		for _, c := range conds {
			if c.Value != nil {
				values = append(values, fmt.Sprintf("%q", c.Value))
			}
		}
		if len(values) == 0 {
			parts = append(parts, fmt.Sprintf("%q has a value", field))
		} else if len(values) == 1 {
			parts = append(parts, fmt.Sprintf("%q is %s", field, values[0]))
		} else {
			parts = append(parts, fmt.Sprintf("%q is one of %s", field, strings.Join(values, ", ")))
		}
	}

	attr.Description += fmt.Sprintf("\n\nOnly applicable when %s.", strings.Join(parts, " and "))
	attr.Description = strings.TrimSpace(attr.Description)
}

// attributesFromDependentSchemas extracts attributes that are conditionally
// visible based on another field's value. The dependentSchemas map keys are
// the "trigger" field names; values use oneOf, if/then, or allOf to describe
// which attributes appear under which conditions.
func attributesFromDependentSchemas(connSchema *jsonschema.Schema) ([]Attribute, error) {
	if len(connSchema.DependentSchemas) == 0 {
		return nil, nil
	}
	var attrs []Attribute
	for triggerField, depSchema := range connSchema.DependentSchemas {
		extracted, err := extractConditionalAttrs(triggerField, depSchema)
		if err != nil {
			return nil, fmt.Errorf("dependentSchemas[%s]: %w", triggerField, err)
		}
		attrs = append(attrs, extracted...)
	}
	return attrs, nil
}

// extractConditionalAttrs handles oneOf, if/then, and allOf structures within
// a single dependentSchemas entry for a given trigger field.
func extractConditionalAttrs(triggerField string, schema *jsonschema.Schema) ([]Attribute, error) {
	var attrs []Attribute

	// oneOf: each entry has properties with the trigger field enum value
	// plus the dependent fields.
	if len(schema.OneOf) > 0 {
		for _, branch := range schema.OneOf {
			if branch.Properties == nil {
				continue
			}
			// Determine the condition value from the trigger field property.
			var condValue interface{}
			if triggerProp, ok := branch.Properties.Get(triggerField); ok {
				if len(triggerProp.Enum) == 1 {
					condValue = triggerProp.Enum[0]
				}
			}
			for pair := branch.Properties.Oldest(); pair != nil; pair = pair.Next() {
				if pair.Key == triggerField {
					continue
				}
				attr, err := tfAttr(pair.Key, pair.Value, branch.Required)
				if err != nil {
					return nil, err
				}
				attr.Conditions = append(attr.Conditions, AttrCondition{
					Field:    triggerField,
					Value:    condValue,
					Required: slices.Contains(branch.Required, pair.Key),
				})
				// Conditional fields are always optional/computed in Terraform.
				attr.Required = false
				attr.Optional = true
				attr.Computed = true
				attrs = append(attrs, attr)
			}
		}
	}

	// if/then: condition on trigger field, then-clause has dependent fields.
	if schema.If != nil && schema.Then != nil {
		condAttrs, err := extractIfThenAttrs(triggerField, schema.If, schema.Then)
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, condAttrs...)
	}

	// allOf: multiple independent if/then conditions.
	for _, entry := range schema.AllOf {
		if entry.If != nil && entry.Then != nil {
			condAttrs, err := extractIfThenAttrs(triggerField, entry.If, entry.Then)
			if err != nil {
				return nil, err
			}
			attrs = append(attrs, condAttrs...)
		}
	}

	return attrs, nil
}

// extractIfThenAttrs handles a single if/then pair, extracting the condition
// value from the if-clause and the dependent fields from the then-clause.
func extractIfThenAttrs(triggerField string, ifSchema, thenSchema *jsonschema.Schema) ([]Attribute, error) {
	var attrs []Attribute

	// Determine condition: could be const-based or contains-based.
	var condValue interface{}
	if ifSchema.Properties != nil {
		if triggerProp, ok := ifSchema.Properties.Get(triggerField); ok {
			if triggerProp.Const != nil {
				condValue = triggerProp.Const
			} else if triggerProp.Contains != nil && triggerProp.Contains.Const != nil {
				condValue = triggerProp.Contains.Const
			}
		}
	}

	if thenSchema.Properties != nil {
		for pair := thenSchema.Properties.Oldest(); pair != nil; pair = pair.Next() {
			if pair.Key == triggerField {
				continue
			}
			attr, err := tfAttr(pair.Key, pair.Value, thenSchema.Required)
			if err != nil {
				return nil, err
			}
			attr.Conditions = append(attr.Conditions, AttrCondition{
				Field:    triggerField,
				Value:    condValue,
				Required: slices.Contains(thenSchema.Required, pair.Key),
			})
			attr.Required = false
			attr.Optional = true
			attr.Computed = true
			attrs = append(attrs, attr)
		}
	}

	// Handle nested if/then within then-clause (multi-condition AND).
	if thenSchema.If != nil && thenSchema.Then != nil {
		nested, err := extractIfThenAttrs(triggerField, thenSchema.If, thenSchema.Then)
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, nested...)
	}

	return attrs, nil
}

func tfAttr(k string, a *jsonschema.Schema, required []string) (Attribute, error) {
	// Check if this is an object with additionalProperties (a map)
	schemaType := a.Type
	if a.Type == "object" {
		// Check if it has additionalProperties of type string
		if a.AdditionalProperties != nil && a.AdditionalProperties.Type == "string" {
			// This is a map[string]string
			schemaType = "map"
		} else if a.Properties == nil || a.Properties.Len() == 0 {
			// Object with no properties defined - treat as a flexible map
			schemaType = "map"
		}
	}

	t, ok := TypeMap[schemaType]
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
		// Only treat as nested object if it has properties (not a map with additionalProperties)
		if schemaType != "map" {
			sa, err := attributesForJSONSchema(a)
			if err != nil {
				return Attribute{}, fmt.Errorf("error inspecting attributes for %s: %w", k, err)
			}
			attr.Attributes = sa
		}
	}
	// Extract enum values. The API returns enums as either plain strings
	// or objects with "value"/"label" keys.
	for _, e := range a.Enum {
		switch v := e.(type) {
		case string:
			attr.EnumValues = append(attr.EnumValues, v)
			attr.EnumLabels = append(attr.EnumLabels, "")
		case map[string]interface{}:
			if val, ok := v["value"].(string); ok {
				attr.EnumValues = append(attr.EnumValues, val)
				label, _ := v["label"].(string)
				attr.EnumLabels = append(attr.EnumLabels, label)
			}
		}
	}
	if len(attr.EnumValues) > 0 {
		hasLabels := false
		for _, l := range attr.EnumLabels {
			if l != "" {
				hasLabels = true
				break
			}
		}
		if hasLabels {
			entries := make([]string, len(attr.EnumValues))
			for i, v := range attr.EnumValues {
				if attr.EnumLabels[i] != "" {
					entries[i] = fmt.Sprintf("<code>%s</code> (%s)", v, attr.EnumLabels[i])
				} else {
					entries[i] = fmt.Sprintf("<code>%s</code>", v)
				}
			}
			attr.Description += fmt.Sprintf(" Valid values: %s.", strings.Join(entries, ", "))
		} else {
			entries := make([]string, len(attr.EnumValues))
			for i, v := range attr.EnumValues {
				entries[i] = fmt.Sprintf("<code>%s</code>", v)
			}
			attr.Description += fmt.Sprintf(" Valid values: %s.", strings.Join(entries, ", "))
		}
		attr.Description = strings.TrimSpace(attr.Description)
	}

	// Add default value to description when present.
	if a.Default != nil {
		attr.Description += fmt.Sprintf(" Default: <code>%v</code>.", a.Default)
		attr.Description = strings.TrimSpace(attr.Description)
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

// displayType maps a Terraform schema attribute type to a human-readable name
// for use in documentation.
func displayType(attrType string) string {
	switch attrType {
	case "schema.StringAttribute":
		return "String"
	case "schema.BoolAttribute":
		return "Boolean"
	case "schema.NumberAttribute", "schema.Int64Attribute":
		return "Number"
	case "schema.SetAttribute":
		return "Set of String"
	case "schema.SingleNestedAttribute":
		return "Attributes"
	case "schema.MapAttribute":
		return "Map of String"
	case "schema.SetNestedAttribute":
		return "Attributes Set"
	default:
		return "String"
	}
}

// hasNestedAttrs returns true if the attribute contains nested attributes
// that should be rendered as a separate section.
func hasNestedAttrs(a Attribute) bool {
	if len(a.Attributes) > 0 {
		return true
	}
	if a.Elem != nil && len(a.Elem.Attributes) > 0 {
		return true
	}
	return false
}

// nestedAttrs returns the nested attributes for an attribute, handling both
// direct nesting (SingleNestedAttribute) and element nesting (SetNestedAttribute).
func nestedAttrs(a Attribute) []Attribute {
	if len(a.Attributes) > 0 {
		return a.Attributes
	}
	if a.Elem != nil {
		return a.Elem.Attributes
	}
	return nil
}

// renderAttrLine renders a single attribute as a markdown list item.
func renderAttrLine(a Attribute, anchorPrefix string) string {
	typeName := displayType(a.AttrType)

	annotations := []string{typeName}
	if a.Sensitive {
		annotations = append(annotations, "Sensitive")
	}
	if a.Required {
		annotations = append(annotations, "Required")
	} else {
		annotations = append(annotations, "Optional")
	}

	desc := strings.TrimSpace(a.Description)

	if hasNestedAttrs(a) {
		anchor := anchorPrefix + "--" + a.AttrName
		if desc != "" {
			desc += fmt.Sprintf(" See [below for nested schema](#%s).", anchor)
		} else {
			desc = fmt.Sprintf("See [below for nested schema](#%s).", anchor)
		}
	}

	if desc != "" {
		return fmt.Sprintf("- `%s` (%s) %s", a.AttrName, strings.Join(annotations, ", "), desc)
	}
	return fmt.Sprintf("- `%s` (%s)", a.AttrName, strings.Join(annotations, ", "))
}

// renderNestedSections recursively renders nested schema sections for all
// attributes that have sub-attributes.
func renderNestedSections(attrs []Attribute, anchorPrefix, pathPrefix string) string {
	var sb strings.Builder
	for _, a := range attrs {
		if !hasNestedAttrs(a) {
			continue
		}
		anchor := anchorPrefix + "--" + a.AttrName
		attrPath := pathPrefix + "." + a.AttrName

		fmt.Fprintf(&sb, "\n<a id=%q></a>\n### Nested Schema for `%s`\n\n", anchor, attrPath)

		nested := nestedAttrs(a)
		for _, na := range nested {
			sb.WriteString(renderAttrLine(na, anchor))
			sb.WriteString("\n")
		}

		// Recurse into nested attributes.
		sub := renderNestedSections(nested, anchor, attrPath)
		if sub != "" {
			sb.WriteString(sub)
		}
	}
	return sb.String()
}

// generateSchemaMarkdown produces the full schema documentation section for a
// connection resource, replacing the default tfplugindocs rendering.
func generateSchemaMarkdown(conn Connection) string {
	var sb strings.Builder

	sb.WriteString("## Schema\n\n")

	// Top-level attributes (same for all connections).
	sb.WriteString("- `name` (String, Required)\n")
	if len(conn.Attributes) > 0 {
		sb.WriteString("- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).\n")
	} else {
		sb.WriteString("- `configuration` (Attributes, Optional)\n")
	}
	sb.WriteString("- `organization` (String, Optional) Organization ID.\n")
	fmt.Fprintf(&sb, "- `id` (String, Read-only) %s Connection identifier.\n", conn.Name)
	forceDestroyDesc := "Indicates whether dependent models, syncs, and bulk syncs should be cascade-deleted when this connection is destroyed."
	if raw, err := os.ReadFile(forceDestroyDescriptionFile); err == nil {
		// Indent continuation lines so they stay inside the markdown list item.
		paragraphs := strings.Split(strings.TrimSpace(string(raw)), "\n\n")
		forceDestroyDesc = strings.Join(paragraphs, "\n\n  ")
	}
	fmt.Fprintf(&sb, "- `force_destroy` (Boolean, Optional) %s\n", forceDestroyDesc)

	// Nested configuration schema.
	if len(conn.Attributes) > 0 {
		sb.WriteString("\n<a id=\"nestedatt--configuration\"></a>\n### Nested Schema for `configuration`\n\n")
		for _, a := range conn.Attributes {
			sb.WriteString(renderAttrLine(a, "nestedatt--configuration"))
			sb.WriteString("\n")
		}
		sub := renderNestedSections(conn.Attributes, "nestedatt--configuration", "configuration")
		if sub != "" {
			sb.WriteString(sub)
		}
	}

	return sb.String()
}

// exampleHeading converts an example filename like "example_with_ssh_tunnel.tf"
// into a heading like "With SSH Tunnel".
func exampleHeading(filename string) string {
	name := strings.TrimSuffix(filename, ".tf")
	name = strings.TrimPrefix(name, "example_")
	name = strings.ReplaceAll(name, "_", " ")
	// Title-case each word.
	words := strings.Fields(name)
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

// generateExamplesSection builds the Example Usage section for the doc
// template. If curated example files (example_*.tf) exist in the resource's
// example directory, only those are rendered (with headings derived from
// filenames). Otherwise, the auto-generated resource.tf is used.
func generateExamplesSection(resourceName string) string {
	exampleDir := filepath.Join(exampleResourceOutputPath,
		fmt.Sprintf("polytomic_%s_connection", resourceName))

	matches, _ := filepath.Glob(filepath.Join(exampleDir, "example_*.tf"))
	if len(matches) == 0 {
		// Fall back to the auto-generated example via tfplugindocs.
		return `{{ tffile .ExampleFile }}`
	}

	slices.Sort(matches)
	var sb strings.Builder
	for _, m := range matches {
		heading := exampleHeading(filepath.Base(m))
		// Path relative to repo root for tffile.
		rel := filepath.Join(exampleDir, filepath.Base(m))
		fmt.Fprintf(&sb, "### %s\n\n{{ tffile %q }}\n\n", heading, rel)
	}
	return strings.TrimSpace(sb.String())
}

func writeConnectionDocTemplate(r Connection) error {
	tmpl, err := template.New("resource_doc.md.tmpl").ParseFiles(connectionResourceDocTemplate)
	if err != nil {
		return fmt.Errorf("error parsing doc template: %w", err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		Name           string
		ResourceName   string
		SchemaMarkdown string
		Examples       string
	}{
		Name:           r.Name,
		ResourceName:   r.ResourceName,
		SchemaMarkdown: generateSchemaMarkdown(r),
		Examples:       generateExamplesSection(r.ResourceName),
	})
	if err != nil {
		return fmt.Errorf("error executing doc template for %s: %w", r.Connection, err)
	}

	outputFile := filepath.Join(docTemplateOutputPath, fmt.Sprintf("%s_connection.md.tmpl", r.ResourceName))
	return os.WriteFile(outputFile, buf.Bytes(), 0644)
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
