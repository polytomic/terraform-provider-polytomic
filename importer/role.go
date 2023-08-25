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

	Resources map[string]polytomic.Role
}

func NewRoles(c *polytomic.Client) *Roles {
	return &Roles{
		c:         c,
		Resources: make(map[string]polytomic.Role),
	}
}

func (r *Roles) Init(ctx context.Context) error {
	roles, err := r.c.Permissions().ListRoles(ctx)
	if err != nil {
		return err
	}

	for _, role := range roles {
		// Skip system roles, they are not editable
		if role.System {
			continue
		}
		name := provider.ValidName(provider.ToSnakeCase(role.Name))
		r.Resources[name] = role
	}

	return nil

}

func (r *Roles) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	for name, role := range r.Resources {
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("resource", []string{RoleResource, name})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(role.Name))
		if role.OrganizationID != "" {
			resourceBlock.Body().SetAttributeValue("organization", cty.StringVal(role.OrganizationID))
		}

		writer.Write(ReplaceRefs(hclFile.Bytes(), refs))
	}
	return nil
}

func (r *Roles) GenerateImports(ctx context.Context, writer io.Writer) error {
	for name, role := range r.Resources {
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			RoleResource,
			name,
			role.ID)))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", role.Name)))
	}
	return nil
}

func (r *Roles) Filename() string {
	return RolesResourceFileName
}

func (r *Roles) ResourceRefs() map[string]string {
	result := make(map[string]string)
	for name, role := range r.Resources {
		result[role.ID] = name
	}
	return result
}

func (r *Roles) DatasourceRefs() map[string]string {
	return nil
}
