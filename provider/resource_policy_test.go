package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
					testAccPolicyExists(name),
				),
			},
		},
	})
}

func testAccPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["polytomic_policy.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_policy.test")
		}
		if rs.Primary.Attributes["name"] != name {
			return fmt.Errorf("name is %s; want %s", rs.Primary.Attributes["name"], name)
		}
		return nil

	}
}

func testAccPolicyResource(name string) string {
	return fmt.Sprintf(`
resource "polytomic_policy" "test" {
	name           = "%s"
	policy_actions = [
		{
			action = "apply_policy"
			role_ids = [
				polytomic_role.test.id
			]
		}
	]
}
resource "polytomic_role" "test" {
	name         = "%s"
}
`, name, name)
}
