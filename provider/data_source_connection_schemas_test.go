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

func TestAccConnectionSchemasDataSource(t *testing.T) {
	name := fmt.Sprintf("TestAccConnectionSchemas-%s", uuid.NewString())
	dataSource := "data.polytomic_connection_schemas.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, connectionSchemasDataSourceTemplate, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					// The CSV connection has one schema, named after the file.
					statecheck.ExpectKnownValue(dataSource,
						tfjsonpath.New("schemas"),
						knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(dataSource,
						tfjsonpath.New("schemas").AtSliceIndex(0).AtMapKey("id"),
						knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSource,
						tfjsonpath.New("schemas").AtSliceIndex(0).AtMapKey("fields"),
						knownvalue.Null()),
				},
			},
			{
				Config: TestCaseTfResource(t, connectionSchemasDataSourceTemplate, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
					Extra:  map[string]any{"IncludeFields": true},
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSource,
						tfjsonpath.New("schemas").AtSliceIndex(0).AtMapKey("fields"),
						knownvalue.SetPartial([]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								"name":           knownvalue.StringExact("EMail"),
								"is_primary_key": knownvalue.Bool(false),
							}),
						})),
				},
			},
		},
	})
}

const connectionSchemasDataSourceTemplate = `
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

data "polytomic_connection_schemas" "test" {
  connection_id  = polytomic_csv_connection.test.id
{{if .Extra.IncludeFields}}
  include_fields = true
{{end}}
{{if not .APIKey}}
  organization   = polytomic_organization.test.id
{{end}}
}
`
