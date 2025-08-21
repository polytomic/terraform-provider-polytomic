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
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyResource(name),
				Check: resource.ComposeTestCheckFunc(
					// Check if the resource exists
					testAccPolicyExists(t, name),
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

func testAccPolicyExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		org, ok := s.RootModule().Resources["polytomic_organization.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_organization.test")
		}

		_, ok = s.RootModule().Resources["polytomic_policy.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_policy.test")
		}

		client := testClient(t, org.Primary.ID)
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

func testAccPolicyResource(name string) string {
	return fmt.Sprintf(`
resource "polytomic_organization" "test" {
  name = "%s"
}

resource "polytomic_policy" "test" {
	name           = "%s"
	organization   = polytomic_organization.test.id
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
}
resource "polytomic_role" "test" {
	name         = "%s"
	organization = polytomic_organization.test.id
}
`, name, name, name)
}
