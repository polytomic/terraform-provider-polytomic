package importer

import (
	"context"
	"fmt"
	"io"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/polytomic/polytomic-go"
	ptclient "github.com/polytomic/polytomic-go/client"
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
	c *ptclient.Client

	Resources map[string]*polytomic.RoleResponse
}

func NewRoles(c *ptclient.Client) *Roles {
	return &Roles{
		c:         c,
		Resources: make(map[string]*polytomic.RoleResponse),
	}
}

func (r *Roles) Init(ctx context.Context) error {
	roles, err := r.c.Permissions.Roles.List(ctx)
	if err != nil {
		return err
	}

	for _, role := range roles.Data {
		// Skip system roles, they are not editable
		if pointer.GetBool(role.System) {
			continue
		}
		name := provider.ValidName(provider.ToSnakeCase(pointer.GetString(role.Name)))
		r.Resources[name] = role
	}

	// Organization variable will be handled centrally

	return nil

}

func (r *Roles) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	for _, name := range sortedKeys(r.Resources) {
		role := r.Resources[name]
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		resourceBlock := body.AppendNewBlock("resource", []string{RoleResource, name})
		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(pointer.GetString(role.Name)))
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

		writer.Write(ReplaceRefs(hclFile.Bytes(), refs))
	}
	return nil
}

func (r *Roles) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, name := range sortedKeys(r.Resources) {
		role := r.Resources[name]
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			RoleResource,
			name,
			pointer.GetString(role.Id))))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", pointer.GetString(role.Name))))
	}
	return nil
}

func (r *Roles) Filename() string {
	return RolesResourceFileName
}

func (r *Roles) ResourceRefs() map[string]string {
	result := make(map[string]string)
	for name, role := range r.Resources {
		result[pointer.GetString(role.Id)] = fmt.Sprintf("polytomic_role.%s.id", name)
	}
	return result
}

func (r *Roles) DatasourceRefs() map[string]string {
	return nil
}

func (r *Roles) Variables() []Variable {
	return nil
}
