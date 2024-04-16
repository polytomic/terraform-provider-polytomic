package importer

import (
	"context"
	"fmt"
	"io"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/polytomic/polytomic-go"
	ptclient "github.com/polytomic/polytomic-go/client"
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
	c *ptclient.Client

	Resources map[string]*polytomic.PolicyResponse
}

func NewPolicies(c *ptclient.Client) *Policies {
	return &Policies{
		c:         c,
		Resources: make(map[string]*polytomic.PolicyResponse),
	}
}

func (p *Policies) Init(ctx context.Context) error {
	policies, err := p.c.Permissions.Policies.List(ctx)
	if err != nil {
		return err
	}

	for _, policy := range policies.Data {
		// Skip system policies, they are not editable
		if pointer.GetBool(policy.System) {
			continue
		}
		hyrdatedPolicy, err := p.c.Permissions.Policies.Get(ctx, pointer.GetString(policy.Id))
		if err != nil {
			return err
		}
		name := provider.ValidName(provider.ToSnakeCase(pointer.GetString(policy.Name)))
		p.Resources[name] = hyrdatedPolicy.Data
	}

	return nil

}

func (p *Policies) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	for _, name := range sortedKeys(p.Resources) {
		policy := p.Resources[name]
		hclFile := hclwrite.NewEmptyFile()
		body := hclFile.Body()

		resourceBlock := body.AppendNewBlock("resource", []string{PolicyResource, name})

		resourceBlock.Body().SetAttributeValue("name", cty.StringVal(pointer.GetString(policy.Name)))
		if policy.OrganizationId != nil {
			resourceBlock.Body().SetAttributeValue("organization", cty.StringVal(pointer.GetString(policy.OrganizationId)))
		}

		var policyActions []map[string]interface{}
		for _, action := range policy.PolicyActions {
			policyActions = append(policyActions, map[string]interface{}{
				"action":   action.Action,
				"role_ids": action.RoleIds,
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
			pointer.GetString(policy.Id))))
		writer.Write([]byte(fmt.Sprintf(" # %s\n", pointer.GetString(policy.Name))))
	}
	return nil
}

func (p *Policies) Filename() string {
	return PoliciesResourceFileName
}

func (p *Policies) ResourceRefs() map[string]string {
	result := make(map[string]string)
	for name, policy := range p.Resources {
		result[pointer.GetString(policy.Id)] = name
	}
	return result
}

func (p *Policies) DatasourceRefs() map[string]string {
	return nil
}

func (p *Policies) Variables() []Variable {
	return nil
}
