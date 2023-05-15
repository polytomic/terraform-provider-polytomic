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

	Resources   []Connection
	Datasources []Connection
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
		c: c,
	}
}

func (c *Connections) Init(ctx context.Context) error {
	conns, err := c.c.Connections().List(ctx)
	if err != nil {
		return err
	}
	for _, conn := range conns {
		if r, ok := provider.ConnectionsMap[conn.Type.ID]; ok {
			resp := &resource.MetadataResponse{}
			r.Metadata(ctx, resource.MetadataRequest{
				ProviderTypeName: provider.Name,
			}, resp)
			c.Resources = append(c.Resources, Connection{
				ID:            conn.ID,
				Resource:      resp.TypeName,
				Name:          conn.Name,
				Organization:  conn.OrganizationId,
				Configuration: conn.Configuration,
			})

		} else if d, ok := provider.ConnectionDatasourcesMap[conn.Type.ID]; ok {
			resp := &datasource.MetadataResponse{}
			d.Metadata(ctx, datasource.MetadataRequest{
				ProviderTypeName: provider.Name,
			}, resp)
			c.Datasources = append(c.Datasources, Connection{
				ID:           conn.ID,
				Resource:     resp.TypeName,
				Name:         conn.Name,
				Organization: conn.OrganizationId,
			})
		} else {
			log.Warn().Msgf("connection type %s not supported", conn.Type.ID)
		}
	}
	return nil
}

func (c *Connections) GenerateTerraformFiles(ctx context.Context, writer io.Writer) error {
	for _, conn := range c.Datasources {
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("data", []string{conn.Resource, provider.ValidName(provider.ToSnakeCase(conn.Name))})
		resourceBlock.Body().SetAttributeValue("id", cty.StringVal(conn.ID))
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(conn.Name))
		resourceBlock.Body().SetAttributeValue("organization", cty.StringVal(conn.Organization))
		body.AppendNewline()

		writer.Write(hclFile.Bytes())
	}

	for _, conn := range c.Resources {
		config := typeConverter(conn.Configuration)
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("resource", []string{conn.Resource, provider.ToSnakeCase(conn.Name)})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(conn.Name))
		resourceBlock.Body().SetAttributeValue("organization", cty.StringVal(conn.Organization))
		resourceBlock.Body().SetAttributeValue("configuration", config)
		body.AppendNewline()

		writer.Write(hclFile.Bytes())
	}
	return nil

}

func (c *Connections) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, conn := range c.Resources {
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			conn.Resource,
			provider.ValidName(provider.ToSnakeCase(conn.Name)),
			conn.ID)))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", conn.Name)))
	}
	return nil
}

func (c *Connections) Filename() string {
	return ConnectionsResourceFileName
}
