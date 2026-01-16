package importer

import (
	"context"
	"fmt"
	"io"

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

var (
	_ Importable = &Connections{}
)

type Connections struct {
	c *ptclient.Client

	Resources   map[string]Connection
	Datasources map[string]Connection
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

			// Filter config to only include fields that exist in the schema
			// and remove sensitive fields
			filteredConfig := make(map[string]interface{})
			missingRequiredFields := []string{}
			for k, v := range config {
				if attr, exists := configSchema.Attributes[k]; exists {
					if !attr.IsSensitive() {
						filteredConfig[k] = v
					}
				}
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

			// Skip OAuth connections that have required sensitive fields
			// These cannot be imported because the credentials are not readable from the API
			if len(missingRequiredFields) > 0 {
				log.Warn().
					Str("connection", pointer.GetString(conn.Name)).
					Str("type", resp.TypeName).
					Strs("missing_fields", missingRequiredFields).
					Msg("skipping connection with required sensitive fields (OAuth connections cannot be imported)")
				continue
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

		writer.Write(hclFile.Bytes())
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
	return nil
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
