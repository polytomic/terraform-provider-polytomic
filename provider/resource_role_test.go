package provider

import (
	"context"
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
					// Check if the resource exists
					testAccRoleExists(name),
					// Check the name
					resource.TestCheckResourceAttr("polytomic_role.test", "name", name),
				),
			},
		},
	})
}

func testAccRoleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources["polytomic_role.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_role.test")
		}

		client := testClient()
		roles, err := client.Permissions().ListRoles(context.TODO())
		if err != nil {
			return err
		}
		var found bool
		for _, role := range roles {
			if role.Name == name {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("role %s not found", name)
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
