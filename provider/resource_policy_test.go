package provider

import (
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPolicy(t *testing.T) {
	name := "TestAccPolicy"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, policyResourceTemplate, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				Check: resource.ComposeTestCheckFunc(
					// Check if the resource exists
					testAccPolicyExists(t, name, APIKey()),
					// Check the name
					resource.TestCheckResourceAttr("polytomic_policy.test", "name", name),
					// Number of policy actions
					resource.TestCheckResourceAttr("polytomic_policy.test", "policy_actions.#", "2"),
					// Check the first policy action
					resource.TestCheckResourceAttr("polytomic_policy.test", "policy_actions.0.action", "apply_policy"),
					resource.TestCheckResourceAttr("polytomic_policy.test", "policy_actions.0.role_ids.#", "1")),
			},
		},
	})
}

func testAccPolicyExists(t *testing.T, name string, apiKey bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var orgID string
		if !apiKey {
			org, ok := s.RootModule().Resources["polytomic_organization.test"]
			if !ok {
				return fmt.Errorf("not found: %s", "polytomic_organization.test")
			}
			orgID = org.Primary.ID
		}

		_, ok := s.RootModule().Resources["polytomic_policy.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_policy.test")
		}

		client := testClient(t, orgID)
		policies, err := client.Permissions.Policies.List(t.Context())
		if err != nil {
			return err
		}
		var found bool
		for _, policy := range policies.Data {
			if pointer.Get(policy.Name) == name {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("policy %s not found in API response", name)
		}

		return nil

	}
}

const policyResourceTemplate = `
{{if not .APIKey}}
resource "polytomic_organization" "test" {
	name = "{{.Name}}"
}
{{end}}

resource "polytomic_policy" "test" {
	name           = "{{.Name}}"
	policy_actions = [
		{
			action = "apply_policy"
			role_ids = [
				polytomic_role.test.id
			]
		},
		{
			action = "delete"
			role_ids = [
				polytomic_role.test.id
			]
		}
	]
{{if not .APIKey}}
	organization   = polytomic_organization.test.id
{{end}}
}

resource "polytomic_role" "test" {
	name         = "{{.Name}}"
{{if not .APIKey}}
	organization   = polytomic_organization.test.id
{{end}}
}`
