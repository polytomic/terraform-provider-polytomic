package importer

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/polytomic/polytomic-go"
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
	c *polytomic.Client

	Resources   map[string]Connection
	Datasources map[string]Connection
}

type Connection struct {
	ID            string
	Type          string
	Resource      string
	Name          string
	Organization  string
	Configuration interface{}
}

func NewConnections(c *polytomic.Client) *Connections {
	return &Connections{
		c:           c,
		Resources:   make(map[string]Connection),
		Datasources: make(map[string]Connection),
	}
}

func (c *Connections) Init(ctx context.Context) error {
	conns, err := c.c.Connections().List(ctx)
	if err != nil {
		return err
	}
	for _, conn := range conns {
		name := provider.ValidName(provider.ToSnakeCase(conn.Name))
		if r, ok := provider.ConnectionsMap[conn.Type.ID]; ok {
			resp := &resource.MetadataResponse{}
			r.Metadata(ctx, resource.MetadataRequest{
				ProviderTypeName: provider.Name,
			}, resp)
			c.Resources[name] = Connection{
				ID:            conn.ID,
				Resource:      resp.TypeName,
				Name:          conn.Name,
				Organization:  conn.OrganizationId,
				Configuration: conn.Configuration,
			}

		} else if d, ok := provider.ConnectionDatasourcesMap[conn.Type.ID]; ok {
			resp := &datasource.MetadataResponse{}
			d.Metadata(ctx, datasource.MetadataRequest{
				ProviderTypeName: provider.Name,
			}, resp)
			c.Datasources[name] = Connection{
				ID:           conn.ID,
				Resource:     resp.TypeName,
				Name:         conn.Name,
				Organization: conn.OrganizationId,
			}
		} else {
			log.Warn().Msgf("connection type %s not supported", conn.Type.ID)
		}
	}
	return nil
}

func (c *Connections) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	for _, name := range sortedKeys(c.Datasources) {
		conn := c.Datasources[name]
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("data", []string{conn.Resource, name})
		resourceBlock.Body().SetAttributeValue("id", cty.StringVal(conn.ID))
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(conn.Name))
		resourceBlock.Body().SetAttributeValue("organization", cty.StringVal(conn.Organization))
		body.AppendNewline()

		writer.Write(ReplaceRefs(hclFile.Bytes(), refs))
	}

	for _, name := range sortedKeys(c.Resources) {
		conn := c.Resources[name]
		config := typeConverter(conn.Configuration)
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("resource", []string{conn.Resource, name})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(conn.Name))
		resourceBlock.Body().SetAttributeValue("organization", cty.StringVal(conn.Organization))
		resourceBlock.Body().SetAttributeValue("configuration", config)
		body.AppendNewline()

		writer.Write(hclFile.Bytes())
	}
	return nil

}

func (c *Connections) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, name := range sortedKeys(c.Resources) {
		conn := c.Resources[name]
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			conn.Resource,
			name,
			conn.ID)))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", conn.Name)))
	}
	return nil
}

func (c *Connections) Filename() string {
	return ConnectionsResourceFileName
}

func (c *Connections) ResourceRefs() map[string]string {
	result := make(map[string]string)
	for name, conn := range c.Resources {
		result[conn.ID] = fmt.Sprintf("%s.%s.id", conn.Resource, name)
	}
	return result
}

func (c *Connections) DatasourceRefs() map[string]string {
	result := make(map[string]string)
	for name, conn := range c.Datasources {
		result[conn.ID] = fmt.Sprintf("%s.%s.id", conn.Resource, name)
	}
	return result
}
