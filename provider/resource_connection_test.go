package provider

import (
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccConnectionResource(t *testing.T) {
	name := fmt.Sprintf("TestAccConnection-%s", uuid.NewString())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, connectionResourceTemplate, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_csv_connection.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"polytomic_csv_connection.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					// Verify resource exists in API
					testAccConnectionExists(t, name, APIKey()),
				),
			},
		},
	})
}

func testAccConnectionExists(t *testing.T, name string, apiKey bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var orgID string
		if !apiKey {
			org, ok := s.RootModule().Resources["polytomic_organization.test"]
			if !ok {
				return fmt.Errorf("not found: %s", "polytomic_organization.test")
			}
			orgID = org.Primary.ID
		}

		resource, ok := s.RootModule().Resources["polytomic_csv_connection.test"]
		if !ok {
			return fmt.Errorf("not found: polytomic_csv_connection.test")
		}

		client := testClient(t, orgID)
		conn, err := client.Connections.Get(t.Context(), resource.Primary.ID)
		if err != nil {
			return err
		}

		if pointer.Get(conn.Data.Name) != name {
			return fmt.Errorf("expected connection name %q, got %q", name, pointer.Get(conn.Data.Name))
		}

		return nil
	}
}

const connectionResourceTemplate = `
{{if not .APIKey}}
resource "polytomic_organization" "test" {
  name = "{{.Name}}"
}
{{end}}
resource "polytomic_csv_connection" "test" {
  name          = "{{.Name}}"
  configuration = {
    url = "https://gist.githubusercontent.com/jpalawaga/20df01c463b82950cc7421e5117a67bc/raw/14bae37fb748114901f7cfdaa5834e4b417537d5/"
  }
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}
`
