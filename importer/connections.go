package importer

import (
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mitchellh/mapstructure"
	ptclient "github.com/polytomic/polytomic-go/v25/client"
	"github.com/polytomic/terraform-provider-polytomic/provider"
	"github.com/rs/zerolog/log"
	"github.com/zclconf/go-cty/cty"
)

const (
	ConnectionsResourceFileName = "connections.tf"
)

// varSentinelRe matches the placeholder strings we plant for required
// sensitive fields. The HCL writer emits these as quoted string values; we
// post-process them into bare var.<name> traversals.
var varSentinelRe = regexp.MustCompile(`"__VARREF_([a-zA-Z0-9_]+)__"`)

// varRefSentinel returns the placeholder string for a Terraform input
// variable reference. It is replaced with `var.<name>` after HCL rendering.
func varRefSentinel(name string) string {
	return fmt.Sprintf("__VARREF_%s__", name)
}

// substituteVarRefs converts placeholder strings emitted by varRefSentinel
// into unquoted var.<name> traversals in the rendered HCL bytes.
func substituteVarRefs(b []byte) []byte {
	return varSentinelRe.ReplaceAll(b, []byte("var.$1"))
}

var (
	_ Importable = &Connections{}
)

type Connections struct {
	c *ptclient.Client

	Resources   map[string]Connection
	Datasources map[string]Connection

	// variables collects the input variables generated for required
	// sensitive fields whose values cannot be read back from the API.
	variables []Variable
}

type Connection struct {
	ID            *string
	Type          *string
	Resource      string
	Name          *string
	Organization  *string
	Configuration interface{}
}

func NewConnections(c *ptclient.Client) *Connections {
	return &Connections{
		c:           c,
		Resources:   make(map[string]Connection),
		Datasources: make(map[string]Connection),
	}
}

func (c *Connections) Init(ctx context.Context) error {
	conns, err := c.c.Connections.List(ctx)
	if err != nil {
		return err
	}
	for _, conn := range conns.Data {
		name := provider.ValidName(provider.ToSnakeCase(pointer.GetString(conn.Name)))
		if r, ok := provider.ConnectionsMap[pointer.GetString(conn.Type.ID)]; ok {
			resp := &resource.MetadataResponse{}
			r.Metadata(ctx, resource.MetadataRequest{
				ProviderTypeName: provider.Name,
			}, resp)

			schemaResp := &resource.SchemaResponse{}
			r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

			var config map[string]interface{}
			err := mapstructure.Decode(conn.Configuration, &config)
			if err != nil {
				return err
			}

			// Normalize configuration keys to snake_case to match schema
			config = normalizeConfigKeys(config)

			configSchema, ok := schemaResp.Schema.Attributes["configuration"].(schema.SingleNestedAttribute)
			if !ok {
				return fmt.Errorf("not single nested attribute %s", resp.TypeName)
			}

			// Filter config to only include fields that exist in the schema,
			// excluding sensitive fields (cannot be read from the API) and
			// computed-only fields (server-managed; the provider rejects them
			// in config).
			filteredConfig := make(map[string]interface{})
			for k, v := range config {
				attr, exists := configSchema.Attributes[k]
				if !exists {
					continue
				}
				if attr.IsSensitive() {
					continue
				}
				if attr.IsComputed() && !attr.IsRequired() && !attr.IsOptional() {
					continue
				}
				filteredConfig[k] = v
			}

			// A connection can hold a value the provider's connection
			// definition does not accept, such as an auth mode the public
			// definition omits. The generated config would fail validation.
			if invalid := invalidConfigValues(ctx, configSchema.Attributes, filteredConfig); len(invalid) > 0 {
				log.Warn().
					Str("connection", pointer.GetString(conn.Name)).
					Str("type", resp.TypeName).
					Strs("invalid_fields", invalid).
					Msg("skipping connection (configuration not accepted by the provider)")
				continue
			}

			// Required fields can be missing after filtering: the API never
			// returns sensitive fields, and a connection can predate a field
			// that is now required.
			missingRequiredFields := []string{}
			unsupportedFields := []string{}
			for _, fieldName := range sortedKeys(configSchema.Attributes) {
				attr := configSchema.Attributes[fieldName]
				if !attr.IsRequired() {
					continue
				}
				if _, exists := filteredConfig[fieldName]; exists {
					continue
				}
				missingRequiredFields = append(missingRequiredFields, fieldName)
				if _, ok := varTypeMap[attr.GetType().String()]; !ok {
					unsupportedFields = append(unsupportedFields, fieldName)
				}
			}

			connTypeID := pointer.GetString(conn.Type.ID)
			if len(missingRequiredFields) > 0 {
				// True OAuth connections cannot be reproduced from a Terraform
				// config — the refresh token only exists after an interactive
				// consent flow. Skip them.
				if provider.OAuthConnections[connTypeID] {
					log.Warn().
						Str("connection", pointer.GetString(conn.Name)).
						Str("type", resp.TypeName).
						Strs("missing_fields", missingRequiredFields).
						Msg("skipping OAuth connection (credentials not retrievable from API)")
					continue
				}
				if len(unsupportedFields) > 0 {
					log.Warn().
						Str("connection", pointer.GetString(conn.Name)).
						Str("type", resp.TypeName).
						Strs("missing_fields", unsupportedFields).
						Msg("skipping connection (missing required fields cannot be input variables)")
					continue
				}

				// For non-OAuth connections, generate input variables for the
				// missing required fields so the user can supply them at
				// apply time.
				for _, fieldName := range missingRequiredFields {
					attr := configSchema.Attributes[fieldName]
					varName := fmt.Sprintf("%s_%s", name, fieldName)
					filteredConfig[fieldName] = varRefSentinel(varName)
					c.variables = append(c.variables, Variable{
						Name:      varName,
						Type:      varTypeMap[attr.GetType().String()],
						Sensitive: attr.IsSensitive(),
					})
				}
				log.Info().
					Str("connection", pointer.GetString(conn.Name)).
					Str("type", resp.TypeName).
					Strs("fields", missingRequiredFields).
					Msg("generating input variables for missing required fields")
			}

			config = filteredConfig

			// Validate the connection resource schema
			validator, err := NewSchemaValidator(ctx, r)
			if err != nil {
				return fmt.Errorf("failed to create schema validator for %s: %w", resp.TypeName, err)
			}

			// Build field mapping for this connection
			mapping := map[string]interface{}{
				"name":          pointer.GetString(conn.Name),
				"organization":  pointer.GetString(conn.OrganizationID),
				"configuration": config,
			}

			// Validate the mapping
			if err := validator.ValidateMapping(mapping); err != nil {
				return fmt.Errorf("schema validation failed for connection '%s' (%s): %w",
					pointer.GetString(conn.Name), resp.TypeName, err)
			}

			c.Resources[name] = Connection{
				ID:            conn.ID,
				Resource:      resp.TypeName,
				Name:          conn.Name,
				Organization:  conn.OrganizationID,
				Configuration: config,
			}

		} else if d, ok := provider.ConnectionDatasourcesMap[pointer.GetString(conn.Type.ID)]; ok {
			resp := &datasource.MetadataResponse{}
			d.Metadata(ctx, datasource.MetadataRequest{
				ProviderTypeName: provider.Name,
			}, resp)

			// Get datasource schema for validation
			schemaReq := datasource.SchemaRequest{}
			schemaResp := &datasource.SchemaResponse{}
			d.Schema(ctx, schemaReq, schemaResp)

			// Build field mapping for this datasource. The data source reads
			// name, so it is not part of the configuration.
			mapping := map[string]interface{}{
				"id":           pointer.GetString(conn.ID),
				"organization": pointer.GetString(conn.OrganizationID),
			}

			// Validate datasource schema by checking that required fields exist
			for fieldName := range mapping {
				if _, exists := schemaResp.Schema.Attributes[fieldName]; !exists {
					log.Warn().Msgf("datasource %s missing expected field '%s'", resp.TypeName, fieldName)
				}
			}

			c.Datasources[name] = Connection{
				ID:           conn.ID,
				Resource:     resp.TypeName,
				Name:         conn.Name,
				Organization: conn.OrganizationID,
			}

		} else {
			log.Warn().Msgf("connection type %s not supported", pointer.GetString(conn.Type.ID))
		}
	}

	// Organization variable will be handled centrally

	return nil
}

func (c *Connections) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	// Check if we should use organization variable
	// useOrgVariable := len(c.organizationIDs) == 1

	for _, name := range sortedKeys(c.Datasources) {
		conn := c.Datasources[name]
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("data", []string{conn.Resource, name})
		resourceBlock.Body().SetAttributeValue("id", cty.StringVal(pointer.GetString(conn.ID)))
		resourceBlock.Body().SetAttributeTraversal("organization",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "organization_id",
				},
			},
		)
		body.AppendNewline()

		writer.Write(hclFile.Bytes())
	}

	for _, name := range sortedKeys(c.Resources) {
		conn := c.Resources[name]
		config := typeConverter(conn.Configuration)
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("resource", []string{conn.Resource, name})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(pointer.GetString(conn.Name)))
		resourceBlock.Body().SetAttributeTraversal("organization",
			hcl.Traversal{
				hcl.TraverseRoot{
					Name: "local",
				},
				hcl.TraverseAttr{
					Name: "organization_id",
				},
			},
		)

		resourceBlock.Body().SetAttributeValue("configuration", config)
		body.AppendNewline()

		writer.Write(substituteVarRefs(hclFile.Bytes()))
	}
	return nil

}

func (c *Connections) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, name := range sortedKeys(c.Resources) {
		conn := c.Resources[name]
		fmt.Fprintf(writer, "terraform import %s.%s %s # %s\n",
			conn.Resource,
			name,
			pointer.Get(conn.ID),
			pointer.Get(conn.Name),
		)
	}
	return nil
}

func (c *Connections) Filename() string {
	return ConnectionsResourceFileName
}

func (c *Connections) ResourceRefs() map[string]string {
	result := make(map[string]string)
	for name, conn := range c.Resources {
		result[pointer.GetString(conn.ID)] = fmt.Sprintf("%s.%s.id", conn.Resource, name)
	}
	return result
}

func (c *Connections) DatasourceRefs() map[string]string {
	result := make(map[string]string)
	for name, conn := range c.Datasources {
		result[pointer.GetString(conn.ID)] = fmt.Sprintf("data.%s.%s.id", conn.Resource, name)
	}
	return result
}

func (c *Connections) Variables() []Variable {
	return c.variables
}

// invalidConfigValues returns the configuration fields whose string values the
// schema's validators reject, formatted as field="value".
func invalidConfigValues(ctx context.Context, attrs map[string]schema.Attribute, config map[string]interface{}) []string {
	var invalid []string
	for _, k := range sortedKeys(config) {
		attr, ok := attrs[k].(schema.StringAttribute)
		if !ok {
			continue
		}
		// Empty strings are omitted from the generated configuration.
		s, ok := config[k].(string)
		if !ok || s == "" {
			continue
		}
		for _, v := range attr.Validators {
			resp := &validator.StringResponse{}
			v.ValidateString(ctx, validator.StringRequest{
				Path:        path.Root("configuration").AtName(k),
				ConfigValue: types.StringValue(s),
			}, resp)
			if resp.Diagnostics.HasError() {
				invalid = append(invalid, fmt.Sprintf("%s=%q", k, s))
				break
			}
		}
	}
	return invalid
}

// normalizeConfigKeys converts configuration keys from camelCase to snake_case
// to match the Terraform provider schema expectations
func normalizeConfigKeys(config map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{})
	for k, v := range config {
		// Only convert if the key contains uppercase letters (is camelCase)
		// If already snake_case, leave it alone to avoid double conversion
		snakeKey := k
		if containsUpperCase(k) {
			snakeKey = provider.ToSnakeCase(k)
		}

		// Recursively normalize nested maps
		if nestedMap, ok := v.(map[string]interface{}); ok {
			normalized[snakeKey] = normalizeConfigKeys(nestedMap)
		} else {
			normalized[snakeKey] = v
		}
	}
	return normalized
}

// containsUpperCase checks if a string contains any uppercase letters
func containsUpperCase(s string) bool {
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			return true
		}
	}
	return false
}
