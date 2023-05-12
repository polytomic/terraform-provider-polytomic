package importer

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/polytomic/polytomic-go"
	"github.com/polytomic/terraform-provider-polytomic/provider"
	"github.com/zclconf/go-cty/cty"
)

const (
	RolesResourceFileName = "roles.tf"
	RoleResource          = "polytomic_role"
)

var (
	_ Importable = &Roles{}
)

type Roles struct {
	c *polytomic.Client

	Resources []polytomic.Role
}

func NewRoles(c *polytomic.Client) *Roles {
	return &Roles{
		c: c,
	}
}

func (p *Roles) Init(ctx context.Context) error {
	roles, err := p.c.Permissions().ListRoles(ctx)
	if err != nil {
		return err
	}

	for _, role := range roles {
		// Skip system roles, they are not editable
		if role.System {
			continue
		}

		p.Resources = append(p.Resources, role)
	}

	return nil

}

func (p *Roles) GenerateTerraformFiles(ctx context.Context, writer io.Writer) error {
	for _, role := range p.Resources {
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		name := provider.ValidNamer(provider.ToSnakeCase(role.Name))

		resourceBlock := body.AppendNewBlock("resource", []string{RoleResource, name})

		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(role.Name))
		if role.OrganizationID != "" {
			resourceBlock.Body().SetAttributeValue("organization", cty.StringVal(role.OrganizationID))
		}

		writer.Write(hclFile.Bytes())
	}
	return nil
}

func (p *Roles) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, role := range p.Resources {
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			RoleResource,
			provider.ValidNamer(provider.ToSnakeCase(role.Name)),
			role.ID)))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", role.Name)))
	}
	return nil
}

func (p *Roles) Filename() string {
	return RolesResourceFileName
}
