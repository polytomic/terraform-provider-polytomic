package provider

import (
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_policy.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"polytomic_policy.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"polytomic_policy.test",
						tfjsonpath.New("policy_actions"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"polytomic_policy.test",
						tfjsonpath.New("policy_actions").AtSliceIndex(0).AtMapKey("action"),
						knownvalue.StringExact("apply_policy"),
					),
					statecheck.ExpectKnownValue(
						"polytomic_policy.test",
						tfjsonpath.New("policy_actions").AtSliceIndex(0).AtMapKey("role_ids"),
						knownvalue.ListSizeExact(1),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccPolicyExists(t, name, APIKey()),
				),
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
