package provider

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccConnectionSchemaDataSource_Basic(t *testing.T) {
	name := fmt.Sprintf("TestAccConnectionSchema-%s", uuid.NewString())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, connectionSchemaDataSourceTemplate, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.polytomic_connection_schema.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("data"),
					),
					statecheck.ExpectKnownValue(
						"data.polytomic_connection_schema.test",
						tfjsonpath.New("connection_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.polytomic_connection_schema.test",
						tfjsonpath.New("schema_id"),
						knownvalue.StringExact("data"),
					),
					statecheck.ExpectKnownValue(
						"data.polytomic_connection_schema.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.polytomic_connection_schema.test", "name", "data"),
					resource.TestCheckResourceAttr("data.polytomic_connection_schema.test", "schema_id", "data"),
					resource.TestCheckResourceAttrSet("data.polytomic_connection_schema.test", "connection_id"),
					resource.TestCheckResourceAttrSet("data.polytomic_connection_schema.test", "fields.#"),
					resource.TestCheckResourceAttrSet("data.polytomic_connection_schema.test", "id"),
					// Verify that the schema contains expected fields from the CSV
					testAccCheckSchemaHasField("data.polytomic_connection_schema.test", "firstname"),
					testAccCheckSchemaHasField("data.polytomic_connection_schema.test", "lastname"),
					testAccCheckSchemaHasField("data.polytomic_connection_schema.test", "EMail"),
					testAccCheckSchemaHasField("data.polytomic_connection_schema.test", "favouritefood"),
				),
			},
		},
	})
}

// testAccCheckSchemaHasField checks if a schema contains a field with the given name
func testAccCheckSchemaHasField(dataSourceName, fieldName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("not found: %s", dataSourceName)
		}

		// Iterate through the fields set to find the field by name
		foundField := false
		for key, value := range rs.Primary.Attributes {
			// Look for fields.X.name attributes that match our field name
			if len(key) > 7 && key[:7] == "fields." && len(key) > 12 && key[len(key)-5:] == ".name" {
				if value == fieldName {
					foundField = true
					break
				}
			}
		}

		if !foundField {
			return fmt.Errorf("field %q not found in schema %s", fieldName, dataSourceName)
		}

		return nil
	}
}

const connectionSchemaDataSourceTemplate = `
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

data "polytomic_connection_schema" "test" {
  connection_id = polytomic_csv_connection.test.id
  schema_id     = "data"
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}
`
