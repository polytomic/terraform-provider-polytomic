package provider

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/polytomic/polytomic-go/v25"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

func TestPrimaryKeyOverrides(t *testing.T) {
	field := func(id string, effective bool, source, override *bool) *polytomic.SchemaField {
		return &polytomic.SchemaField{
			ID:                 pointer.To(id),
			IsPrimaryKey:       pointer.To(effective),
			SourcePrimaryKey:   source,
			PrimaryKeyOverride: override,
		}
	}
	yes, no := pointer.To(true), pointer.To(false)
	fields := []*polytomic.SchemaField{
		field("id", true, yes, nil),         // detected key
		field("email", false, no, nil),      // plain field
		field("legacy", true, no, yes),      // key forced on by an override
		field("disabled", false, yes, no),   // detected key forced off
		field("unused", false, no, no),      // redundant override
		field("created_at", false, no, nil), // plain field
	}

	for _, tc := range []struct {
		name   string
		fields []*polytomic.SchemaField
		want   []string
		result map[string]bool
	}{
		{
			name:   "replaces the detected key",
			fields: fields,
			want:   []string{"email"},
			result: map[string]bool{"id": false, "email": true, "legacy": false, "disabled": false, "unused": false},
		},
		{
			name:   "keeps the detected key when listed",
			fields: fields,
			want:   []string{"id", "email"},
			result: map[string]bool{"id": true, "email": true, "legacy": false, "disabled": false, "unused": false},
		},
		{
			name:   "unmarks effective keys when provenance is not reported",
			fields: []*polytomic.SchemaField{field("id", true, nil, nil), field("email", false, nil, nil)},
			want:   []string{"email"},
			result: map[string]bool{"id": false, "email": true},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			overrides, err := primaryKeyOverrides(tc.fields, tc.want)
			if err != nil {
				t.Fatal(err)
			}
			got := make(map[string]bool, len(overrides))
			for _, o := range overrides {
				got[o.FieldID] = o.IsPrimaryKey
			}
			if !reflect.DeepEqual(got, tc.result) {
				t.Errorf("got %v, want %v", got, tc.result)
			}
		})
	}

	t.Run("rejects unknown fields", func(t *testing.T) {
		_, err := primaryKeyOverrides(fields, []string{"email", "nope", "missing"})
		if err == nil || err.Error() != "no such fields: missing, nope" {
			t.Errorf("got %v", err)
		}
	})
}

func TestEffectivePrimaryKeys(t *testing.T) {
	got := effectivePrimaryKeys([]*polytomic.SchemaField{
		{ID: pointer.To("tenant_id"), IsPrimaryKey: pointer.To(true)},
		{ID: pointer.To("name")},
		{ID: pointer.To("order_no"), IsPrimaryKey: pointer.To(true)},
	})
	if want := []string{"order_no", "tenant_id"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestAccConnectionSchemaPrimaryKeys_ReplacesDetectedKey pins a Postgres table's
// key to a column other than its declared primary key, and checks that a reset
// made outside Terraform is reported as drift.
func TestAccConnectionSchemaPrimaryKeys_ReplacesDetectedKey(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("%s must be set for acceptance tests", resource.EnvTfAcc)
	}
	name := fmt.Sprintf("TestAccSchemaPK-%s", uuid.NewString())
	config := TestCaseTfResource(t, connectionSchemaPrimaryKeysPostgresTemplate, TestCaseTfArgs{
		Name:     name,
		APIKey:   APIKey(),
		Postgres: testPostgresConfig(t),
	})
	const pk = "polytomic_connection_schema_primary_keys.test"
	const schemaID = "polytomic.sync_test_other"
	var organization, connectionID string

	keyChecks := []statecheck.StateCheck{
		statecheck.ExpectKnownValue(pk, tfjsonpath.New("field_ids"),
			knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("label")})),
		statecheck.ExpectKnownValue("data.polytomic_connection_schema.test",
			tfjsonpath.New("fields_by_id").AtMapKey("id"),
			knownvalue.ObjectPartial(map[string]knownvalue.Check{
				"is_primary_key":       knownvalue.Bool(false),
				"source_primary_key":   knownvalue.Bool(true),
				"primary_key_override": knownvalue.Bool(false),
			})),
		statecheck.ExpectKnownValue("data.polytomic_connection_schema.test",
			tfjsonpath.New("fields_by_id").AtMapKey("label"),
			knownvalue.ObjectPartial(map[string]knownvalue.Check{
				"is_primary_key":       knownvalue.Bool(true),
				"source_primary_key":   knownvalue.Bool(false),
				"primary_key_override": knownvalue.Bool(true),
			})),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:            config,
				ConfigStateChecks: keyChecks,
				Check: func(s *terraform.State) error {
					rs, ok := s.RootModule().Resources[pk]
					if !ok {
						return fmt.Errorf("not found: %s", pk)
					}
					organization = rs.Primary.Attributes["organization"]
					connectionID = rs.Primary.Attributes["connection_id"]
					return nil
				},
			},
			{
				// Resetting the overrides outside Terraform restores the
				// detected key, which the next plan must undo.
				PreConfig: func() {
					testAccResetPrimaryKeys(t, organization, connectionID, schemaID)
				},
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:            config,
				ConfigStateChecks: keyChecks,
			},
		},
	})
}

func testAccResetPrimaryKeys(t *testing.T, organization, connectionID, schemaID string) {
	t.Helper()
	ctx := context.Background()
	p, err := providerclient.NewClientProvider(providerclient.OptionsFromEnv())
	if err != nil {
		t.Fatal(err)
	}
	client, err := p.Client(ctx, organization)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Schemas.ResetPrimaryKeys(ctx, connectionID, schemaID); err != nil {
		t.Fatal(err)
	}
}

const connectionSchemaPrimaryKeysPostgresTemplate = `
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

# sync_test_other declares id as its primary key; replace it with label.
resource "polytomic_connection_schema_primary_keys" "test" {
  connection_id = polytomic_postgresql_connection.test.id
  schema_id     = "polytomic.sync_test_other"
  field_ids     = ["label"]
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}

data "polytomic_connection_schema" "test" {
  connection_id = polytomic_postgresql_connection.test.id
  schema_id     = "polytomic.sync_test_other"
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
  depends_on = [polytomic_connection_schema_primary_keys.test]
}
`

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
