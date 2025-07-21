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

			configSchema, ok := schemaResp.Schema.Attributes["configuration"].(schema.SingleNestedAttribute)
			if !ok {
				return fmt.Errorf("not single nested attribute %s", resp.TypeName)
			}
			for k, v := range configSchema.Attributes {
				if _, ok := config[k]; ok {
					if v.IsSensitive() {
						delete(config, k)
					}
				}
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
