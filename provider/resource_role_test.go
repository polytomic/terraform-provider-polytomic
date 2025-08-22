package provider

import (
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
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, policyResourceTemplate, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				Check: resource.ComposeTestCheckFunc(
					// Check if the resource exists
					testAccRoleExists(t, name, APIKey()),
					// Check the name
					resource.TestCheckResourceAttr("polytomic_role.test", "name", name),
				),
			},
		},
	})
}

func testAccRoleExists(t *testing.T, name string, apiKey bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var orgID string
		if !apiKey {
			org, ok := s.RootModule().Resources["polytomic_organization.test"]
			if !ok {
				return fmt.Errorf("not found: %s", "polytomic_organization.test")
			}
			orgID = org.Primary.ID
		}

		_, ok := s.RootModule().Resources["polytomic_role.test"]
		if !ok {
			return fmt.Errorf("not found: %s", "polytomic_role.test")
		}

		client := testClient(t, orgID)
		roles, err := client.Permissions.Roles.List(t.Context())
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

const roleResourceTemplate = `
{{if not .APIKey}}
resource "polytomic_organization" "test" {
	name = "{{.Name}}"
}
{{end}}

resource "polytomic_role" "test" {
	name         = "{{.Name}}"
	{{if not .APIKey}}
	organization = polytomic_organization.test.id
	{{end}}
}
`
