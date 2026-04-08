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

func TestAccSyncResource(t *testing.T) {
	name := fmt.Sprintf("TestAccSync-%s", uuid.NewString())
	postgres := testPostgresConfig(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, syncResourceTemplate, TestCaseTfArgs{
					Name:     name,
					APIKey:   APIKey(),
					Postgres: postgres,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_sync.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"polytomic_sync.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"polytomic_sync.test",
						tfjsonpath.New("mode"),
						knownvalue.StringExact("replace"),
					),
					statecheck.ExpectKnownValue(
						"polytomic_sync.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(false),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, APIKey()),
				),
			},
		},
	})
}

func testAccSyncExists(t *testing.T, name string, apiKey bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var orgID string
		if !apiKey {
			org, ok := s.RootModule().Resources["polytomic_organization.test"]
			if !ok {
				return fmt.Errorf("not found: %s", "polytomic_organization.test")
			}
			orgID = org.Primary.ID
		}

		resource, ok := s.RootModule().Resources["polytomic_sync.test"]
		if !ok {
			return fmt.Errorf("not found: polytomic_sync.test")
		}

		client := testClient(t, orgID)
		sync, err := client.ModelSync.Get(t.Context(), resource.Primary.ID)
		if err != nil {
			return err
		}

		if pointer.Get(sync.Data.Name) != name {
			return fmt.Errorf("expected sync name %q, got %q", name, pointer.Get(sync.Data.Name))
		}

		return nil
	}
}

const syncResourceTemplate = `
{{if not .APIKey}}
resource "polytomic_organization" "test" {
  name = "{{.Name}}"
}
{{end}}

resource "polytomic_postgresql_connection" "test" {
  name = "{{.Name}}-postgres"
  configuration = {
    hostname = "{{.Postgres.Host}}"
    database = "{{.Postgres.Database}}"
    username = "{{.Postgres.Username}}"
    password = "{{.Postgres.Password}}"
    port     = {{.Postgres.Port}}
  }
{{if not .APIKey}}
  organization = polytomic_organization.test.id
{{end}}
}

resource "polytomic_model" "test" {
  name          = "{{.Name}}-model"
  configuration = jsonencode({
    "query" = "SELECT email FROM polytomic.sync_test_source"
  })
  connection_id = polytomic_postgresql_connection.test.id
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}

resource "polytomic_sync" "test" {
  name   = "{{.Name}}"
  mode   = "replace"
  active = false
  schedule = {
    frequency = "manual"
  }
  fields = [
    {
      source = {
        field    = "email"
        model_id = polytomic_model.test.id
      }
      target = "email"
    }
  ]
  target = {
    connection_id = polytomic_postgresql_connection.test.id
    object        = "polytomic.sync_test_target"
  }
{{if not .APIKey}}
  organization = polytomic_organization.test.id
{{end}}
}
`
