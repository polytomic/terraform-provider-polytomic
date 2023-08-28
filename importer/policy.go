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
	PoliciesResourceFileName = "policies.tf"
	PolicyResource           = "polytomic_policy"
)

var (
	_ Importable = &Policies{}
)

type Policies struct {
	c *polytomic.Client

	Resources map[string]*polytomic.Policy
}

func NewPolicies(c *polytomic.Client) *Policies {
	return &Policies{
		c:         c,
		Resources: make(map[string]*polytomic.Policy),
	}
}

func (p *Policies) Init(ctx context.Context) error {
	policies, err := p.c.Permissions().ListPolicies(ctx)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		// Skip system policies, they are not editable
		if policy.System {
			continue
		}
		hyrdatedPolicy, err := p.c.Permissions().GetPolicy(ctx, policy.ID)
		if err != nil {
			return err
		}
		name := provider.ValidName(provider.ToSnakeCase(policy.Name))
		p.Resources[name] = hyrdatedPolicy
	}

	return nil

}

func (p *Policies) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	for _, name := range sortedKeys(p.Resources) {
		policy := p.Resources[name]
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()

		resourceBlock := body.AppendNewBlock("resource", []string{PolicyResource, name})

		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(policy.Name))
		if policy.OrganizationID != "" {
			resourceBlock.Body().SetAttributeValue("organization", cty.StringVal(policy.OrganizationID))
		}

		var policyActions []map[string]interface{}
		for _, action := range policy.PolicyActions {
			policyActions = append(policyActions, map[string]interface{}{
				"action":   action.Action,
				"role_ids": action.RoleIDs,
			})
		}
		resourceBlock.Body().SetAttributeValue("policy_actions", typeConverter(policyActions))

		writer.Write(hclFile.Bytes())
	}
	return nil
}

func (p *Policies) GenerateImports(ctx context.Context, writer io.Writer) error {
	for _, name := range sortedKeys(p.Resources) {
		policy := p.Resources[name]
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			PolicyResource,
			name,
			policy.ID)))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", policy.Name)))
	}
	return nil
}

func (p *Policies) Filename() string {
	return PoliciesResourceFileName
}

func (p *Policies) ResourceRefs() map[string]string {
	result := make(map[string]string)
	for name, policy := range p.Resources {
		result[policy.ID] = name
	}
	return result
}

func (p *Policies) DatasourceRefs() map[string]string {
	return nil
}
