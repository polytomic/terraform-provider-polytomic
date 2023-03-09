package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRole(t *testing.T) {
	name := "TestAccRole"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleResource(name),
				Check: resource.ComposeTestCheckFunc(
					testAccRoleExists(name),
				),
			},
		},
	})
}

func testAccRoleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["polytomic_role.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_role.test")
		}
		if rs.Primary.Attributes["name"] != name {
			return fmt.Errorf("name is %s; want %s", rs.Primary.Attributes["name"], name)
		}
		return nil

	}
}

func testAccRoleResource(name string) string {
	return fmt.Sprintf(`
resource "polytomic_role" "test" {
	name         = "%s"
}
`, name)
}
