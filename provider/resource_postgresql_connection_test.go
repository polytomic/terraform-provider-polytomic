package provider

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccPostgresqlConnectionResource exercises a connection type with a
// sensitive configuration field (password) end-to-end: Create, then the
// framework's post-apply refresh+plan (Read) to check for drift. The Polytomic
// API masks sensitive values in responses; the provider preserves the
// user-supplied value via resetSensitiveValues so state stays correct and
// terraform doesn't see the masked response as drift.
func TestAccPostgresqlConnectionResource(t *testing.T) {
	name := fmt.Sprintf("TestAccPGConn-%s", uuid.NewString())
	pg := testPostgresConfig(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create: passes only if the provider echoes the password back into state.
			// The framework also runs refresh+plan after apply and fails on drift,
			// so the Read path is implicitly covered.
			{
				Config: TestCaseTfResource(t, postgresqlConnectionTemplate, TestCaseTfArgs{
					Name:     name,
					APIKey:   APIKey(),
					Postgres: pg,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_postgresql_connection.test",
						tfjsonpath.New("configuration").AtMapKey("password"),
						knownvalue.StringExact(pg.Password),
					),
					statecheck.ExpectKnownValue(
						"polytomic_postgresql_connection.test",
						tfjsonpath.New("configuration").AtMapKey("hostname"),
						knownvalue.StringExact(pg.Host),
					),
				},
			},
		},
	})
}

const postgresqlConnectionTemplate = `
{{if not .APIKey}}
resource "polytomic_organization" "test" {
  name = "{{.Name}}"
}
{{end}}

resource "polytomic_postgresql_connection" "test" {
  name = "{{.Name}}"
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
`
