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

func TestAccConnectionSchemaPrimaryKeys_Basic(t *testing.T) {
	name := fmt.Sprintf("TestAccSchemaPK-%s", uuid.NewString())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, connectionSchemaPrimaryKeysTemplate, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_connection_schema_primary_keys.test",
						tfjsonpath.New("connection_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"polytomic_connection_schema_primary_keys.test",
						tfjsonpath.New("schema_id"),
						knownvalue.StringExact("data"),
					),
					statecheck.ExpectKnownValue(
						"polytomic_connection_schema_primary_keys.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("polytomic_connection_schema_primary_keys.test", "schema_id", "data"),
					resource.TestCheckResourceAttrSet("polytomic_connection_schema_primary_keys.test", "connection_id"),
					resource.TestCheckResourceAttrSet("polytomic_connection_schema_primary_keys.test", "id"),
					// Verify we set exactly 1 field as primary key (EMail)
					resource.TestCheckResourceAttr("polytomic_connection_schema_primary_keys.test", "field_ids.#", "1"),
					// Verify the primary key is the EMail field
					testAccCheckPrimaryKeyIsEmailField("polytomic_connection_schema_primary_keys.test"),
				),
			},
		},
	})
}

func TestAccConnectionSchemaPrimaryKeys_Update(t *testing.T) {
	name := fmt.Sprintf("TestAccSchemaPK-%s", uuid.NewString())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: Set EMail as primary key
				Config: TestCaseTfResource(t, connectionSchemaPrimaryKeysTemplate, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("polytomic_connection_schema_primary_keys.test", "field_ids.#", "1"),
					testAccCheckPrimaryKeyIsEmailField("polytomic_connection_schema_primary_keys.test"),
				),
			},
			{
				// Step 2: Update to use firstname and lastname as composite primary key
				Config: TestCaseTfResource(t, connectionSchemaPrimaryKeysTemplateMultiple, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("polytomic_connection_schema_primary_keys.test", "field_ids.#", "2"),
					testAccCheckPrimaryKeysIncludeFields("polytomic_connection_schema_primary_keys.test", []string{"firstname", "lastname"}),
				),
			},
		},
	})
}

// testAccCheckPrimaryKeyIsEmailField verifies that the primary key field is the EMail field
func testAccCheckPrimaryKeyIsEmailField(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		// Get the connection_id and schema_id to look up the data source
		connectionID := rs.Primary.Attributes["connection_id"]

		// Find the data source to get field information
		dataSourceName := "data.polytomic_connection_schema.test"
		ds, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", dataSourceName)
		}

		// Get the field_ids from the primary keys resource
		var pkFieldID string
		for key, value := range rs.Primary.Attributes {
			if key == "field_ids.0" {
				pkFieldID = value
				break
			}
		}

		if pkFieldID == "" {
			return fmt.Errorf("no field_ids found in primary keys resource")
		}

		// Find the field with this ID in the data source and verify it's EMail
		var fieldName string
		for key, value := range ds.Primary.Attributes {
			// Look for fields.X.id that matches our pkFieldID
			if len(key) > 9 && key[len(key)-3:] == ".id" && value == pkFieldID {
				// Get the corresponding name
				nameKey := key[:len(key)-3] + ".name"
				fieldName = ds.Primary.Attributes[nameKey]
				break
			}
		}

		if fieldName != "EMail" {
			return fmt.Errorf("expected primary key field to be 'EMail', got '%s' (connection: %s)", fieldName, connectionID)
		}

		return nil
	}
}

// testAccCheckPrimaryKeysIncludeFields verifies that the primary keys include the specified fields
func testAccCheckPrimaryKeysIncludeFields(resourceName string, expectedFields []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		// Find the data source to get field information
		dataSourceName := "data.polytomic_connection_schema.test"
		ds, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", dataSourceName)
		}

		// Get all field_ids from the primary keys resource
		pkFieldIDs := make(map[string]bool)
		for key, value := range rs.Primary.Attributes {
			if len(key) > 10 && key[:10] == "field_ids." && key[len(key)-1] != '#' {
				pkFieldIDs[value] = true
			}
		}

		// Map field IDs to names using the data source
		foundFields := make(map[string]bool)
		for key, value := range ds.Primary.Attributes {
			// Look for fields.X.id
			if len(key) > 9 && key[len(key)-3:] == ".id" {
				if pkFieldIDs[value] {
					// Get the corresponding name
					nameKey := key[:len(key)-3] + ".name"
					fieldName := ds.Primary.Attributes[nameKey]
					foundFields[fieldName] = true
				}
			}
		}

		// Check all expected fields are present
		for _, expectedField := range expectedFields {
			if !foundFields[expectedField] {
				return fmt.Errorf("expected primary key field '%s' not found; found: %v", expectedField, foundFields)
			}
		}

		return nil
	}
}

const connectionSchemaPrimaryKeysTemplate = `
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

# Set EMail as the primary key
resource "polytomic_connection_schema_primary_keys" "test" {
  connection_id = polytomic_csv_connection.test.id
  schema_id     = "data"
  field_ids = [
    [for field in data.polytomic_connection_schema.test.fields : field.id if field.name == "EMail"][0]
  ]
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}
`

const connectionSchemaPrimaryKeysTemplateMultiple = `
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

# Set firstname and lastname as composite primary key
resource "polytomic_connection_schema_primary_keys" "test" {
  connection_id = polytomic_csv_connection.test.id
  schema_id     = "data"
  field_ids = [
    [for field in data.polytomic_connection_schema.test.fields : field.id if field.name == "firstname"][0],
    [for field in data.polytomic_connection_schema.test.fields : field.id if field.name == "lastname"][0]
  ]
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}
`
