package importer

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/hcl/v2/hclwrite"
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

	Resources []Connection
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
		r, ok := provider.ConnectionsMap[conn.Type.ID]
		if !ok {
			log.Warn().Msgf("connection type %s not supported", conn.Type.ID)
			continue
		}
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
	}
	return nil
}

func (c *Connections) GenerateTerraformFiles(ctx context.Context, writer io.Writer) error {
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
			provider.ToSnakeCase(conn.Name),
			conn.ID)))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", conn.Name)))
	}
	return nil
}

func (c *Connections) Filename() string {
	return ConnectionsResourceFileName
}
