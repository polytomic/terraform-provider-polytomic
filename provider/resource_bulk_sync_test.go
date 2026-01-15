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

func TestAccBulkSyncResource(t *testing.T) {
	name := fmt.Sprintf("TestAccBulkSync-%s", uuid.NewString())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, bulkSyncResourceTemplate, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_bulk_sync.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"polytomic_bulk_sync.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"polytomic_bulk_sync.test",
						tfjsonpath.New("mode"),
						knownvalue.StringExact("replicate"),
					),
					statecheck.ExpectKnownValue(
						"polytomic_bulk_sync.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name, APIKey()),
				),
			},
		},
	})
}

func testAccBulkSyncExists(t *testing.T, name string, apiKey bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var orgID string
		if !apiKey {
			org, ok := s.RootModule().Resources["polytomic_organization.test"]
			if !ok {
				return fmt.Errorf("not found: %s", "polytomic_organization.test")
			}
			orgID = org.Primary.ID
		}

		resource, ok := s.RootModule().Resources["polytomic_bulk_sync.test"]
		if !ok {
			return fmt.Errorf("not found: polytomic_bulk_sync.test")
		}

		client := testClient(t, orgID)
		sync, err := client.BulkSync.Get(t.Context(), resource.Primary.ID, nil)
		if err != nil {
			return err
		}

		if pointer.Get(sync.Data.Name) != name {
			return fmt.Errorf("expected bulk sync name %q, got %q", name, pointer.Get(sync.Data.Name))
		}

		return nil
	}
}

const bulkSyncResourceTemplate = `
{{if not .APIKey}}
resource "polytomic_organization" "test" {
  name = "{{.Name}}"
}
{{end}}

resource "polytomic_postgresql_connection" "source" {
  name = "{{.Name}}-source"
  configuration = {
    hostname = "localhost"
    database = "source"
    username = "test"
    password = "test"
    port     = 5432
  }
{{if not .APIKey}}
  organization = polytomic_organization.test.id
{{end}}
}

resource "polytomic_postgresql_connection" "dest" {
  name = "{{.Name}}-dest"
  configuration = {
    hostname = "localhost"
    database = "dest"
    username = "test"
    password = "test"
    port     = 5432
  }
{{if not .APIKey}}
  organization = polytomic_organization.test.id
{{end}}
}

resource "polytomic_bulk_sync" "test" {
  name   = "{{.Name}}"
  active = true
  mode   = "replicate"

  schedule = {
    frequency = "manual"
  }

  source = {
    connection_id = polytomic_postgresql_connection.source.id
  }

  destination = {
    connection_id = polytomic_postgresql_connection.dest.id
    configuration = {
      "schema" = "public"
    }
  }

  schemas = []

  discovery = {
    enabled = false
  }

{{if not .APIKey}}
  organization = polytomic_organization.test.id
{{end}}
}
`
