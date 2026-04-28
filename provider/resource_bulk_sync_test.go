package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	texttemplate "text/template"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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

// bulkSyncTestConnectionIDs holds pre-created connection IDs shared across all
// bulk sync acceptance tests. This avoids creating new PostgreSQL connections
// per test, which exhausts the connection pool.
type bulkSyncTestConnectionIDs struct {
	SourceID string
	DestID   string
}

var (
	sharedBulkSyncConns     *bulkSyncTestConnectionIDs
	sharedBulkSyncConnsOnce sync.Once
	sharedBulkSyncConnsErr  error
)

// getSharedBulkSyncConnections creates (or reuses) a pair of PostgreSQL
// connections for bulk sync tests. The connections are created once and shared
// across all tests in the package. They are cleaned up via
// POLYTOMIC_BULK_SYNC_TEST_SOURCE_ID / POLYTOMIC_BULK_SYNC_TEST_DEST_ID env
// vars if pre-existing connections are preferred.
func getSharedBulkSyncConnections(t *testing.T) bulkSyncTestConnectionIDs {
	t.Helper()

	// Allow overriding with pre-existing connection IDs
	if src := os.Getenv("POLYTOMIC_BULK_SYNC_TEST_SOURCE_ID"); src != "" {
		dest := os.Getenv("POLYTOMIC_BULK_SYNC_TEST_DEST_ID")
		require.NotEmpty(t, dest, "POLYTOMIC_BULK_SYNC_TEST_DEST_ID must be set when POLYTOMIC_BULK_SYNC_TEST_SOURCE_ID is set")
		return bulkSyncTestConnectionIDs{SourceID: src, DestID: dest}
	}

	sharedBulkSyncConnsOnce.Do(func() {
		client := testClient(t, "")
		ctx := context.Background()
		postgres := testPostgresConfig(t)

		// Clean up stale shared connections from prior test runs
		conns, err := client.Connections.List(ctx)
		if err == nil {
			for _, c := range conns.Data {
				if strings.HasPrefix(pointer.Get(c.Name), "TestAccBulkSync-shared-") {
					_ = client.Connections.Remove(ctx, pointer.Get(c.Id), &polytomic.ConnectionsRemoveRequest{Force: pointer.ToBool(true)})
				}
			}
		}

		source, err := client.Connections.Create(ctx, &polytomic.CreateConnectionRequestSchema{
			Name: fmt.Sprintf("TestAccBulkSync-shared-%s-source", uuid.NewString()),
			Type: "postgresql",
			Configuration: map[string]any{
				"hostname": postgres.Host,
				"database": postgres.Database,
				"username": postgres.Username,
				"password": postgres.Password,
				"port":     postgres.Port,
			},
		})
		if err != nil {
			sharedBulkSyncConnsErr = fmt.Errorf("creating shared source connection: %w", err)
			return
		}

		dest, err := client.Connections.Create(ctx, &polytomic.CreateConnectionRequestSchema{
			Name: fmt.Sprintf("TestAccBulkSync-shared-%s-dest", uuid.NewString()),
			Type: "postgresql",
			Configuration: map[string]any{
				"hostname": postgres.Host,
				"database": postgres.Database,
				"username": postgres.Username,
				"password": postgres.Password,
				"port":     postgres.Port,
			},
		})
		if err != nil {
			sharedBulkSyncConnsErr = fmt.Errorf("creating shared dest connection: %w", err)
			return
		}

		sharedBulkSyncConns = &bulkSyncTestConnectionIDs{
			SourceID: pointer.Get(source.Data.Id),
			DestID:   pointer.Get(dest.Data.Id),
		}
	})

	require.NoError(t, sharedBulkSyncConnsErr, "failed to create shared bulk sync connections")
	require.NotNil(t, sharedBulkSyncConns, "shared bulk sync connections not initialized")
	return *sharedBulkSyncConns
}

func TestBulkSyncFiltersToSDK(t *testing.T) {
	tests := map[string]struct {
		input       []bulkSyncFilter
		expected    []*polytomic.BulkFilter
		expectError bool
	}{
		"string value": {
			input: []bulkSyncFilter{
				{
					FieldId:  types.StringValue("createdAt"),
					Function: types.StringValue("RelativeOnOrAfter"),
					Value:    jsontypes.NewNormalizedValue(`"48 hours ago"`),
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
		"null value": {
			input: []bulkSyncFilter{
				{
					FieldId:  types.StringValue("status"),
					Function: types.StringValue("IsNull"),
					Value:    jsontypes.NewNormalizedNull(),
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
		"array value": {
			input: []bulkSyncFilter{
				{
					FieldId:  types.StringValue("amount"),
					Function: types.StringValue("Between"),
					Value:    jsontypes.NewNormalizedValue(`[100,200]`),
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
		"number value": {
			input: []bulkSyncFilter{
				{
					FieldId:  types.StringValue("count"),
					Function: types.StringValue("GreaterThan"),
					Value:    jsontypes.NewNormalizedValue(`42`),
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
		"empty input": {
			input:    []bulkSyncFilter{},
			expected: []*polytomic.BulkFilter{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, diags := bulkSyncFiltersToSDK(tc.input)
			if tc.expectError {
				require.True(t, diags.HasError())
				return
			}
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
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
		input         []*polytomic.BulkFilter
		expectedValue []jsontypes.Normalized
	}{
		"string value": {
			input: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("createdAt"),
					Function: polytomic.FilterFunction("RelativeOnOrAfter"),
					Value:    "48 hours ago",
				},
			},
			expectedValue: []jsontypes.Normalized{
				jsontypes.NewNormalizedValue(`"48 hours ago"`),
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
			expectedValue: []jsontypes.Normalized{
				jsontypes.NewNormalizedNull(),
			},
		},
		"array value": {
			input: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("amount"),
					Function: polytomic.FilterFunction("Between"),
					Value:    []interface{}{float64(100), float64(200)},
				},
			},
			expectedValue: []jsontypes.Normalized{
				jsontypes.NewNormalizedValue(`[100,200]`),
			},
		},
		"number value": {
			input: []*polytomic.BulkFilter{
				{
					FieldId:  pointer.ToString("count"),
					Function: polytomic.FilterFunction("GreaterThan"),
					Value:    42,
				},
			},
			expectedValue: []jsontypes.Normalized{
				jsontypes.NewNormalizedValue(`42`),
			},
		},
		"empty input": {
			input:         []*polytomic.BulkFilter{},
			expectedValue: []jsontypes.Normalized{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := bulkSyncFiltersFromSDK(tc.input)
			require.NoError(t, err)
			require.Equal(t, len(tc.expectedValue), len(result))
			for i := range tc.expectedValue {
				assert.Equal(t, tc.expectedValue[i], result[i].Value)
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

	roundTripped, diags := bulkSyncFiltersToSDK(tfFilters)
	require.False(t, diags.HasError())

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
	conns := getSharedBulkSyncConnections(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bulkSyncBasicTestConfig(t, bulkSyncBasicTestArgs{
					Name:               name,
					SourceConnectionID: conns.SourceID,
					DestConnectionID:   conns.DestID,
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
					testAccBulkSyncExists(t, name),
				),
			},
		},
	})
}

func testAccBulkSyncExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources["polytomic_bulk_sync.test"]
		if !ok {
			return fmt.Errorf("not found: polytomic_bulk_sync.test")
		}

		client := testClient(t, "")
		sync, err := client.BulkSync.Get(t.Context(), r.Primary.ID, nil)
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
	conns := getSharedBulkSyncConnections(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:               name,
					SourceConnectionID: conns.SourceID,
					DestConnectionID:   conns.DestID,
					Mode:               "replicate",
					Active:             "true",
					Schemas: `[{
    id = "polytomic.sync_test_source"
    filters = [{
      field_id = "created_at"
      function = "OnOrAfter"
      value    = jsonencode("2024-01-01")
    }]
  }]`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"polytomic_bulk_sync.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name),
				),
			},
		},
	})
}

type bulkSyncBasicTestArgs struct {
	Name               string
	SourceConnectionID string
	DestConnectionID   string
}

func bulkSyncBasicTestConfig(t *testing.T, args bulkSyncBasicTestArgs) string {
	t.Helper()
	tmpl := texttemplate.Must(texttemplate.New("bulk-sync-basic").Parse(bulkSyncResourceTemplate))
	var buf strings.Builder
	require.NoError(t, tmpl.Execute(&buf, args))
	return buf.String()
}

const bulkSyncResourceTemplate = `
resource "polytomic_bulk_sync" "test" {
  name   = "{{.Name}}"
  active = true
  mode   = "replicate"

  schedule = {
    frequency = "manual"
  }

  source = {
    connection_id = "{{.SourceConnectionID}}"
  }

  destination = {
    connection_id = "{{.DestConnectionID}}"
    configuration = jsonencode({
      "schema" = "public"
    })
  }

  schemas = []
}
`

// ---------------------------------------------------------------------------
// Flexible template and helpers for advanced bulk sync tests
// ---------------------------------------------------------------------------

type bulkSyncAdvancedTestArgs struct {
	Name                       string
	SourceConnectionID         string
	DestConnectionID           string
	Mode                       string
	Active                     string // "true" or "false"
	Schemas                    string // Raw HCL for schemas (empty = use empty list)
	AutomaticallyAddNewObjects string // "all", "none", etc. (empty = omit)
	AutomaticallyAddNewFields  string
	ConcurrencyLimit           string // integer as string (empty = omit)
	NormalizeNames             string // "enabled", "disabled", "legacy" (empty = omit)
	DisableRecordTimestamps    string // "true" or "false" (empty = omit)
	DataCutoffTimestamp        string // RFC3339 timestamp (empty = omit)
}

func bulkSyncAdvancedTestConfig(t *testing.T, args bulkSyncAdvancedTestArgs) string {
	t.Helper()
	tmpl := texttemplate.Must(texttemplate.New("bulk-sync-advanced").Parse(bulkSyncAdvancedTestTemplate))
	var buf strings.Builder
	require.NoError(t, tmpl.Execute(&buf, args))
	return buf.String()
}

const bulkSyncAdvancedTestTemplate = `
resource "polytomic_bulk_sync" "test" {
  name   = "{{.Name}}"
  active = {{.Active}}
  mode   = "{{.Mode}}"

  schedule = {
    frequency = "manual"
  }

  source = {
    connection_id = "{{.SourceConnectionID}}"
  }

  destination = {
    connection_id = "{{.DestConnectionID}}"
    configuration = jsonencode({
      "schema" = "public"
    })
  }

{{- if .Schemas}}
  schemas = {{.Schemas}}
{{- else}}
  schemas = []
{{- end}}
{{- if .AutomaticallyAddNewObjects}}
  automatically_add_new_objects = "{{.AutomaticallyAddNewObjects}}"
{{- end}}
{{- if .AutomaticallyAddNewFields}}
  automatically_add_new_fields = "{{.AutomaticallyAddNewFields}}"
{{- end}}
{{- if .ConcurrencyLimit}}
  concurrency_limit = {{.ConcurrencyLimit}}
{{- end}}
{{- if .NormalizeNames}}
  normalize_names = "{{.NormalizeNames}}"
{{- end}}
{{- if .DisableRecordTimestamps}}
  disable_record_timestamps = {{.DisableRecordTimestamps}}
{{- end}}
{{- if .DataCutoffTimestamp}}
  data_cutoff_timestamp = "{{.DataCutoffTimestamp}}"
{{- end}}
}
`

// ---------------------------------------------------------------------------
// Test: Snapshot mode
// ---------------------------------------------------------------------------

// TestAccBulkSyncResourceSnapshot is skipped because the PostgreSQL destination
// does not support "snapshot" mode. Snapshot mode requires a destination type
// like S3 or cloud storage that supports full-table snapshots.
func TestAccBulkSyncResourceSnapshot(t *testing.T) {
	t.Skip("Skipped: PostgreSQL destination does not support snapshot mode")
}

// ---------------------------------------------------------------------------
// Test: Auto-discovery options
// ---------------------------------------------------------------------------

func TestAccBulkSyncResourceAutoDiscovery(t *testing.T) {
	name := fmt.Sprintf("TestAccBulkSyncDisc-%s", uuid.NewString())
	conns := getSharedBulkSyncConnections(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:                       name,
					SourceConnectionID:         conns.SourceID,
					DestConnectionID:           conns.DestID,
					Mode:                       "replicate",
					Active:                     "true",
					AutomaticallyAddNewObjects: "all",
					AutomaticallyAddNewFields:  "all",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("automatically_add_new_objects"),
						knownvalue.StringExact("all"),
					),
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("automatically_add_new_fields"),
						knownvalue.StringExact("all"),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Update lifecycle (modify attributes after creation)
// ---------------------------------------------------------------------------

func TestAccBulkSyncResourceUpdateLifecycle(t *testing.T) {
	name := fmt.Sprintf("TestAccBulkSyncUpd-%s", uuid.NewString())
	conns := getSharedBulkSyncConnections(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with replicate, active=true.
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:               name,
					SourceConnectionID: conns.SourceID,
					DestConnectionID:   conns.DestID,
					Mode:               "replicate",
					Active:             "true",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name),
				),
			},
			// Step 2: Deactivate and set auto-discovery.
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:                       name,
					SourceConnectionID:         conns.SourceID,
					DestConnectionID:           conns.DestID,
					Mode:                       "replicate",
					Active:                     "false",
					AutomaticallyAddNewObjects: "all",
					AutomaticallyAddNewFields:  "onlyIncremental",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("automatically_add_new_objects"),
						knownvalue.StringExact("all"),
					),
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("automatically_add_new_fields"),
						knownvalue.StringExact("onlyIncremental"),
					),
				},
			},
			// Step 3: Re-activate and change auto-discovery settings.
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:                       name,
					SourceConnectionID:         conns.SourceID,
					DestConnectionID:           conns.DestID,
					Mode:                       "replicate",
					Active:                     "true",
					AutomaticallyAddNewObjects: "none",
					AutomaticallyAddNewFields:  "all",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("active"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("automatically_add_new_objects"),
						knownvalue.StringExact("none"),
					),
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("automatically_add_new_fields"),
						knownvalue.StringExact("all"),
					),
				},
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: ImportState
// ---------------------------------------------------------------------------

func TestAccBulkSyncResourceImport(t *testing.T) {
	name := fmt.Sprintf("TestAccBulkSyncImp-%s", uuid.NewString())
	conns := getSharedBulkSyncConnections(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:               name,
					SourceConnectionID: conns.SourceID,
					DestConnectionID:   conns.DestID,
					Mode:               "replicate",
					Active:             "true",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name),
				),
			},
			{
				ResourceName:            "polytomic_bulk_sync.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"destination", "source", "schemas"},
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Concurrency limits and normalize_names
// ---------------------------------------------------------------------------

func TestAccBulkSyncResourceOptions(t *testing.T) {
	name := fmt.Sprintf("TestAccBulkSyncOpts-%s", uuid.NewString())
	conns := getSharedBulkSyncConnections(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:               name,
					SourceConnectionID: conns.SourceID,
					DestConnectionID:   conns.DestID,
					Mode:               "replicate",
					Active:             "true",
					ConcurrencyLimit:   "2",
					NormalizeNames:     "enabled",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("concurrency_limit"),
						knownvalue.Int64Exact(2),
					),
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("normalize_names"),
						knownvalue.StringExact("enabled"),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Schemas with field configuration
// ---------------------------------------------------------------------------

func TestAccBulkSyncResourceSchemaFields(t *testing.T) {
	name := fmt.Sprintf("TestAccBulkSyncFields-%s", uuid.NewString())
	conns := getSharedBulkSyncConnections(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:               name,
					SourceConnectionID: conns.SourceID,
					DestConnectionID:   conns.DestID,
					Mode:               "replicate",
					Active:             "true",
					Schemas: `[{
    id      = "polytomic.sync_test_source"
    enabled = true
    fields = [{
      id        = "email"
      enabled   = true
      obfuscate = false
    }]
  }]`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("schemas"),
						knownvalue.SetSizeExact(1),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Multiple schemas
// ---------------------------------------------------------------------------

func TestAccBulkSyncResourceMultipleSchemas(t *testing.T) {
	name := fmt.Sprintf("TestAccBulkSyncMulti-%s", uuid.NewString())
	conns := getSharedBulkSyncConnections(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:               name,
					SourceConnectionID: conns.SourceID,
					DestConnectionID:   conns.DestID,
					Mode:               "replicate",
					Active:             "true",
					Schemas: `[
    {
      id      = "polytomic.sync_test_source"
      enabled = true
    },
    {
      id      = "polytomic.sync_test_other"
      enabled = true
    }
  ]`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("schemas"),
						knownvalue.SetSizeExact(2),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: disable_record_timestamps
// ---------------------------------------------------------------------------

func TestAccBulkSyncResourceDisableRecordTimestamps(t *testing.T) {
	name := fmt.Sprintf("TestAccBulkSyncDRT-%s", uuid.NewString())
	conns := getSharedBulkSyncConnections(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:                    name,
					SourceConnectionID:      conns.SourceID,
					DestConnectionID:        conns.DestID,
					Mode:                    "replicate",
					Active:                  "true",
					DisableRecordTimestamps: "true",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("disable_record_timestamps"),
						knownvalue.Bool(true),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: data_cutoff_timestamp (top-level)
// ---------------------------------------------------------------------------

func TestAccBulkSyncResourceDataCutoffTimestamp(t *testing.T) {
	t.Skip("Skipped: PostgreSQL source does not support data_cutoff_timestamp")

	name := fmt.Sprintf("TestAccBulkSyncCutoff-%s", uuid.NewString())
	conns := getSharedBulkSyncConnections(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:                name,
					SourceConnectionID:  conns.SourceID,
					DestConnectionID:    conns.DestID,
					Mode:                "replicate",
					Active:              "true",
					DataCutoffTimestamp: "2025-01-01T00:00:00Z",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("data_cutoff_timestamp"),
						knownvalue.NotNull(),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Schema with tracking_field and partition_key
// ---------------------------------------------------------------------------

func TestAccBulkSyncResourceSchemaTrackingField(t *testing.T) {
	name := fmt.Sprintf("TestAccBulkSyncTrack-%s", uuid.NewString())
	conns := getSharedBulkSyncConnections(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
					Name:               name,
					SourceConnectionID: conns.SourceID,
					DestConnectionID:   conns.DestID,
					Mode:               "replicate",
					Active:             "true",
					Schemas: `[{
    id             = "polytomic.sync_test_source"
    enabled        = true
    tracking_field = "updated_at"
  }]`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_bulk_sync.test",
						tfjsonpath.New("schemas"),
						knownvalue.SetSizeExact(1),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccBulkSyncExists(t, name),
				),
			},
		},
	})
}
