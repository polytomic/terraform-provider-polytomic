package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	polytomic "github.com/polytomic/polytomic-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkSyncFiltersToSDK(t *testing.T) {
	tests := map[string]struct {
		input       []bulkSyncFilter
		expected    []*polytomic.BulkFilter
		expectError bool
	}{
		"plain string value": {
			input: []bulkSyncFilter{
				{
					FieldId:   types.StringValue("createdAt"),
					Function:  types.StringValue("RelativeOnOrAfter"),
					Value:     types.StringValue("48 hours ago"),
					ValueJSON: types.StringNull(),
				},
			},
			expected: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("createdAt"),
					Function: polytomic.FilterFunction("RelativeOnOrAfter"),
					Value:    "48 hours ago",
				},
			},
		},
		"string that looks like a number stays a string": {
			input: []bulkSyncFilter{
				{
					FieldId:   types.StringValue("code"),
					Function:  types.StringValue("Equality"),
					Value:     types.StringValue("42"),
					ValueJSON: types.StringNull(),
				},
			},
			expected: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("code"),
					Function: polytomic.FilterFunction("Equality"),
					Value:    "42",
				},
			},
		},
		"null value": {
			input: []bulkSyncFilter{
				{
					FieldId:   types.StringValue("status"),
					Function:  types.StringValue("IsNull"),
					Value:     types.StringNull(),
					ValueJSON: types.StringNull(),
				},
			},
			expected: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("status"),
					Function: polytomic.FilterFunction("IsNull"),
					Value:    nil,
				},
			},
		},
		"array via value_json": {
			input: []bulkSyncFilter{
				{
					FieldId:   types.StringValue("amount"),
					Function:  types.StringValue("Between"),
					Value:     types.StringNull(),
					ValueJSON: types.StringValue(`[100,200]`),
				},
			},
			expected: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("amount"),
					Function: polytomic.FilterFunction("Between"),
					Value:    []interface{}{float64(100), float64(200)},
				},
			},
		},
		"number via value_json": {
			input: []bulkSyncFilter{
				{
					FieldId:   types.StringValue("count"),
					Function:  types.StringValue("GreaterThan"),
					Value:     types.StringNull(),
					ValueJSON: types.StringValue("42"),
				},
			},
			expected: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("count"),
					Function: polytomic.FilterFunction("GreaterThan"),
					Value:    float64(42),
				},
			},
		},
		"invalid value_json": {
			input: []bulkSyncFilter{
				{
					FieldId:   types.StringValue("bad"),
					Function:  types.StringValue("Equality"),
					Value:     types.StringNull(),
					ValueJSON: types.StringValue("not json"),
				},
			},
			expectError: true,
		},
		"empty input": {
			input:    []bulkSyncFilter{},
			expected: []*polytomic.BulkFilter{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := bulkSyncFiltersToSDK(tc.input)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, len(tc.expected), len(result))
			for i := range tc.expected {
				assert.Equal(t, tc.expected[i].FieldId, result[i].FieldId)
				assert.Equal(t, tc.expected[i].Function, result[i].Function)
				assert.Equal(t, tc.expected[i].Value, result[i].Value)
			}
		})
	}
}

func TestBulkSyncFiltersFromSDK(t *testing.T) {
	tests := map[string]struct {
		input             []*polytomic.BulkFilter
		expectedValue     []types.String
		expectedValueJSON []types.String
	}{
		"string value populates value field": {
			input: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("createdAt"),
					Function: polytomic.FilterFunction("RelativeOnOrAfter"),
					Value:    "48 hours ago",
				},
			},
			expectedValue: []types.String{
				types.StringValue("48 hours ago"),
			},
			expectedValueJSON: []types.String{
				types.StringNull(),
			},
		},
		"nil value": {
			input: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("status"),
					Function: polytomic.FilterFunction("IsNull"),
					Value:    nil,
				},
			},
			expectedValue: []types.String{
				types.StringNull(),
			},
			expectedValueJSON: []types.String{
				types.StringNull(),
			},
		},
		"array value populates value_json field": {
			input: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("amount"),
					Function: polytomic.FilterFunction("Between"),
					Value:    []interface{}{float64(100), float64(200)},
				},
			},
			expectedValue: []types.String{
				types.StringNull(),
			},
			expectedValueJSON: []types.String{
				types.StringValue("[100,200]"),
			},
		},
		"number value populates value_json field": {
			input: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("count"),
					Function: polytomic.FilterFunction("GreaterThan"),
					Value:    42,
				},
			},
			expectedValue: []types.String{
				types.StringNull(),
			},
			expectedValueJSON: []types.String{
				types.StringValue("42"),
			},
		},
		"empty input": {
			input:             []*polytomic.BulkFilter{},
			expectedValue:     []types.String{},
			expectedValueJSON: []types.String{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := bulkSyncFiltersFromSDK(tc.input)
			require.NoError(t, err)
			require.Equal(t, len(tc.expectedValue), len(result))
			for i := range tc.expectedValue {
				assert.Equal(t, tc.expectedValue[i], result[i].Value)
				assert.Equal(t, tc.expectedValueJSON[i], result[i].ValueJSON)
			}
		})
	}
}

func TestBulkSyncFiltersRoundTrip(t *testing.T) {
	sdkFilters := []*polytomic.BulkFilter{
		{
			FieldId:  pointer.ToString("createdAt"),
			Function: polytomic.FilterFunction("RelativeOnOrAfter"),
			Value:    "48 hours ago",
		},
		{
			FieldId:  pointer.ToString("amount"),
			Function: polytomic.FilterFunction("Between"),
			Value:    []interface{}{float64(100), float64(200)},
		},
		{
			FieldId:  pointer.ToString("count"),
			Function: polytomic.FilterFunction("GreaterThan"),
			Value:    float64(42),
		},
	}

	tfFilters, err := bulkSyncFiltersFromSDK(sdkFilters)
	require.NoError(t, err)

	roundTripped, err := bulkSyncFiltersToSDK(tfFilters)
	require.NoError(t, err)

	require.Equal(t, len(sdkFilters), len(roundTripped))
	for i := range sdkFilters {
		assert.Equal(t, sdkFilters[i].FieldId, roundTripped[i].FieldId)
		assert.Equal(t, sdkFilters[i].Function, roundTripped[i].Function)
		assert.Equal(t, sdkFilters[i].Value, roundTripped[i].Value)
	}
}

func TestBulkSyncSchemasFromSDK(t *testing.T) {
	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := map[string]struct {
		input    []*polytomic.BulkSchema
		validate func(t *testing.T, result []bulkSyncSchema)
	}{
		"schema with filters and fields": {
			input: []*polytomic.BulkSchema{
				{
					Id:      pointer.ToString("orders"),
					Enabled: pointer.ToBool(true),
					Fields: []*polytomic.BulkField{
						{
							Id:             pointer.ToString("id"),
							Enabled:        pointer.ToBool(true),
							Obfuscated:     pointer.ToBool(false),
							OutputName:     pointer.ToString("id"),
							UserOutputName: pointer.ToString(""),
						},
					},
					Filters: []*polytomic.BulkFilter{
						{
							FieldId:  pointer.ToString("createdAt"),
							Function: polytomic.FilterFunction("RelativeOnOrAfter"),
							Value:    "48 hours ago",
						},
					},
				},
			},
			validate: func(t *testing.T, result []bulkSyncSchema) {
				require.Len(t, result, 1)
				s := result[0]
				assert.Equal(t, "orders", s.Id.ValueString())
				assert.Equal(t, true, s.Enabled.ValueBool())
				assert.False(t, s.Fields.IsNull())
				assert.False(t, s.Filters.IsNull())
				assert.Equal(t, 1, len(s.Fields.Elements()))
				assert.Equal(t, 1, len(s.Filters.Elements()))
			},
		},
		"schema without filters or fields": {
			input: []*polytomic.BulkSchema{
				{
					Id:      pointer.ToString("users"),
					Enabled: pointer.ToBool(false),
				},
			},
			validate: func(t *testing.T, result []bulkSyncSchema) {
				require.Len(t, result, 1)
				s := result[0]
				assert.Equal(t, "users", s.Id.ValueString())
				assert.True(t, s.Fields.IsNull())
				assert.True(t, s.Filters.IsNull())
			},
		},
		"schema with data cutoff timestamp": {
			input: []*polytomic.BulkSchema{
				{
					Id:                  pointer.ToString("events"),
					DataCutoffTimestamp: &ts,
				},
			},
			validate: func(t *testing.T, result []bulkSyncSchema) {
				require.Len(t, result, 1)
				assert.False(t, result[0].DataCutoffTimestamp.IsNull())
			},
		},
		"schema with nil data cutoff timestamp": {
			input: []*polytomic.BulkSchema{
				{
					Id: pointer.ToString("events"),
				},
			},
			validate: func(t *testing.T, result []bulkSyncSchema) {
				require.Len(t, result, 1)
				assert.True(t, result[0].DataCutoffTimestamp.IsNull())
			},
		},
		"obfuscated maps to obfuscate": {
			input: []*polytomic.BulkSchema{
				{
					Id: pointer.ToString("sensitive"),
					Fields: []*polytomic.BulkField{
						{
							Id:         pointer.ToString("ssn"),
							Obfuscated: pointer.ToBool(true),
						},
					},
				},
			},
			validate: func(t *testing.T, result []bulkSyncSchema) {
				require.Len(t, result, 1)
				var fields []bulkSyncSchemaField
				diags := result[0].Fields.ElementsAs(t.Context(), &fields, false)
				require.False(t, diags.HasError())
				require.Len(t, fields, 1)
				assert.Equal(t, true, fields[0].Obfuscate.ValueBool())
			},
		},
		"empty input": {
			input: []*polytomic.BulkSchema{},
			validate: func(t *testing.T, result []bulkSyncSchema) {
				assert.Len(t, result, 0)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, diags := bulkSyncSchemasFromSDK(t.Context(), tc.input)
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			tc.validate(t, result)
		})
	}
}

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

func TestAccBulkSyncResourceWithFilters(t *testing.T) {
	name := fmt.Sprintf("TestAccBulkSyncFilters-%s", uuid.NewString())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, bulkSyncResourceWithFiltersTemplate, TestCaseTfArgs{
					Name:   name,
					APIKey: APIKey(),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_bulk_sync.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name, APIKey()),
				),
			},
		},
	})
}

const bulkSyncResourceWithFiltersTemplate = `
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

  schemas = [{
    id = "public.users"
    filters = [{
      field_id = "created_at"
      function = "OnOrAfter"
      value    = "2024-01-01"
    }]
  }]

  discovery = {
    enabled = false
  }

{{if not .APIKey}}
  organization = polytomic_organization.test.id
{{end}}
}
`

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
