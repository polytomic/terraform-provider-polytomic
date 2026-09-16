package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	polytomic "github.com/polytomic/polytomic-go/v25"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncDataFromResponseOverrideFields(t *testing.T) {
	tests := map[string]struct {
		response string
		expected []overrideField
	}{
		"reported in override_fields": {
			response: `{
				"fields": [
					{"source": {"model_id": "0b9a3a5e-6c1f-4d1e-9f0a-2d6c1b1e4a01", "field": "email"}, "target": "email"}
				],
				"override_fields": [
					{"target": "name", "override_value": "default"},
					{"target": "status", "override_value": "active", "new": true, "sync_mode": "create"}
				]
			}`,
			expected: []overrideField{
				{
					Target:        types.StringValue("name"),
					New:           types.BoolNull(),
					OverrideValue: types.StringValue("default"),
					SyncMode:      types.StringNull(),
				},
				{
					Target:        types.StringValue("status"),
					New:           types.BoolValue(true),
					OverrideValue: types.StringValue("active"),
					SyncMode:      types.StringValue("create"),
				},
			},
		},
		"none": {
			response: `{
				"fields": [
					{"source": {"model_id": "0b9a3a5e-6c1f-4d1e-9f0a-2d6c1b1e4a01", "field": "email"}, "target": "email"}
				]
			}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := t.Context()
			var sync polytomic.ModelSyncV5Response
			require.NoError(t, json.Unmarshal([]byte(tc.response), &sync))
			sync.Target = &polytomic.ModelSyncV5Target{ConnectionID: "conn", Object: pointer.ToString("contacts")}

			data, diags := syncDataFromResponse(ctx, &sync)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			assert.Len(t, data.Fields.Elements(), 1)

			if tc.expected == nil {
				assert.True(t, data.OverrideFields.IsNull(), "override_fields should be null")
				return
			}
			var overrides []overrideField
			require.False(t, data.OverrideFields.ElementsAs(ctx, &overrides, false).HasError())
			assert.ElementsMatch(t, tc.expected, overrides)
		})
	}
}

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
