package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					testAccRoleExists(t, name),
					// Check the name
					resource.TestCheckResourceAttr("polytomic_role.test", "name", name),
				),
			},
		},
	})
}

func testAccRoleExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources["polytomic_role.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_role.test")
		}

		client := testClient(t)
		roles, err := client.Permissions.Roles.List(context.TODO())
		if err != nil {
			return err
		}
		var found bool
		for _, role := range roles.Data {
			if pointer.Get(role.Name) == name {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("role %s not found in API response", name)
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
