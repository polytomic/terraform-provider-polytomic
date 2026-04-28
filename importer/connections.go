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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/mitchellh/mapstructure"
	ptclient "github.com/polytomic/polytomic-go/client"
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
		if r, ok := provider.ConnectionsMap[pointer.GetString(conn.Type.Id)]; ok {
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
			missingRequiredFields := []string{}
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

			// Check if any required fields are missing after filtering
			for fieldName, attr := range configSchema.Attributes {
				if attr.IsRequired() {
					if _, exists := filteredConfig[fieldName]; !exists {
						if attr.IsSensitive() {
							missingRequiredFields = append(missingRequiredFields, fieldName)
						}
					}
				}
			}

			connTypeID := pointer.GetString(conn.Type.Id)
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

				// For non-OAuth connections, generate input variables for the
				// missing required sensitive fields so the user can supply
				// them at apply time.
				for _, fieldName := range missingRequiredFields {
					varName := fmt.Sprintf("%s_%s", name, fieldName)
					filteredConfig[fieldName] = varRefSentinel(varName)
					c.variables = append(c.variables, Variable{
						Name:      varName,
						Type:      "string",
						Sensitive: true,
					})
				}
				log.Info().
					Str("connection", pointer.GetString(conn.Name)).
					Str("type", resp.TypeName).
					Strs("fields", missingRequiredFields).
					Msg("generating input variables for required sensitive fields")
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
				"organization":  pointer.GetString(conn.OrganizationId),
				"configuration": config,
			}

			// Validate the mapping
			if err := validator.ValidateMapping(mapping); err != nil {
				return fmt.Errorf("schema validation failed for connection '%s' (%s): %w",
					pointer.GetString(conn.Name), resp.TypeName, err)
			}

			c.Resources[name] = Connection{
				ID:            conn.Id,
				Resource:      resp.TypeName,
				Name:          conn.Name,
				Organization:  conn.OrganizationId,
				Configuration: config,
			}

		} else if d, ok := provider.ConnectionDatasourcesMap[pointer.GetString(conn.Type.Id)]; ok {
			resp := &datasource.MetadataResponse{}
			d.Metadata(ctx, datasource.MetadataRequest{
				ProviderTypeName: provider.Name,
			}, resp)

			// Get datasource schema for validation
			schemaReq := datasource.SchemaRequest{}
			schemaResp := &datasource.SchemaResponse{}
			d.Schema(ctx, schemaReq, schemaResp)

			// Build field mapping for this datasource
			mapping := map[string]interface{}{
				"id":           pointer.GetString(conn.Id),
				"name":         pointer.GetString(conn.Name),
				"organization": pointer.GetString(conn.OrganizationId),
			}

			// Validate datasource schema by checking that required fields exist
			for fieldName := range mapping {
				if _, exists := schemaResp.Schema.Attributes[fieldName]; !exists {
					log.Warn().Msgf("datasource %s missing expected field '%s'", resp.TypeName, fieldName)
				}
			}

			c.Datasources[name] = Connection{
				ID:           conn.Id,
				Resource:     resp.TypeName,
				Name:         conn.Name,
				Organization: conn.OrganizationId,
			}

		} else {
			log.Warn().Msgf("connection type %s not supported", pointer.GetString(conn.Type.Id))
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
