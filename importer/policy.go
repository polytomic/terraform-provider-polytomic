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

	Resources []*polytomic.Policy
}

func NewPolicies(c *polytomic.Client) *Policies {
	return &Policies{
		c: c,
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

		p.Resources = append(p.Resources, hyrdatedPolicy)
	}

	return nil

}

func (p *Policies) GenerateTerraformFiles(ctx context.Context, writer io.Writer) error {
	for _, policy := range p.Resources {
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()
		name := provider.ValidName(provider.ToSnakeCase(policy.Name))

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
	for _, policy := range p.Resources {
		writer.Write([]byte(fmt.Sprintf("terraform import %s.%s %s",
			PolicyResource,
			provider.ValidName(provider.ToSnakeCase(policy.Name)),
			policy.ID)))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", policy.Name)))
	}
	return nil
}

func (p *Policies) Filename() string {
	return PoliciesResourceFileName
}
