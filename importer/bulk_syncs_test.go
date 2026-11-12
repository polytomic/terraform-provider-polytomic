package importer

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/polytomic/polytomic-go/v25"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func TestBulkSyncExportPreservesConfiguration(t *testing.T) {
	b := NewBulkSyncs(newTestClient(t, map[string]string{
		"/api/bulk/syncs/sync-1/schemas":          `{"data":[{"id":"contacts","enabled":true},{"id":"unselected","enabled":false}]}`,
		"/api/bulk/syncs/sync-1/schemas/contacts": `{"data":{"id":"contacts","enabled":true,"excluded_fields":["secret"],"history_enabled":true,"destination_history_output_name":"contacts_history","fields":[{"id":"email","enabled":true},{"id":"secret","enabled":false}]}}`,
	}))
	now := time.Now()
	source := map[string]any{"replication_slot": "", "empty": []any{}, "nullable": nil, "nested": map[string]any{"empty": "", "heterogeneous": []any{"x", true, nil, []any{float64(1)}}}}
	destination := map[string]any{"schema": "public", "advanced": map[string]any{"table_prefix": "", "enabled": false}}
	b.Resources["test"] = &polytomic.BulkSyncResponse{
		ID: pointer.ToString("sync-1"), Name: pointer.ToString("Test"), Active: pointer.ToBool(false), Mode: pointer.To(polytomic.BulkSyncTargetModeReplicate),
		SourceConnectionID: pointer.ToString("source"), DestinationConnectionID: pointer.ToString("destination"),
		SourceConfiguration: source, DestinationConfiguration: destination,
		DefaultSchedule: &polytomic.BulkSyncDefaultScheduleResponse{ID: pointer.ToString("schedule-1"), Frequency: polytomic.ScheduleFrequencyHourly, Minute: pointer.ToString("15"), CreatedAt: &now, UpdatedAt: &now},
	}
	var output bytes.Buffer
	require.NoError(t, b.GenerateTerraformFiles(t.Context(), &output, nil))
	file, diags := hclsyntax.ParseConfig(output.Bytes(), "bulk_syncs.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors(), "%s", diags)
	body := file.Body.(*hclsyntax.Body).Blocks[0].Body
	ctx := &hcl.EvalContext{Functions: map[string]function.Function{"jsonencode": stdlib.JSONEncodeFunc}}
	for name, want := range map[string]map[string]any{"source": source, "destination": destination} {
		value, diags := body.Attributes[name].Expr.Value(ctx)
		require.False(t, diags.HasErrors(), "%s", diags)
		expected, err := json.Marshal(want)
		require.NoError(t, err)
		require.JSONEq(t, string(expected), value.AsValueMap()["configuration"].AsString())
	}
	schedule, diags := body.Attributes["schedule"].Expr.Value(ctx)
	require.False(t, diags.HasErrors())
	attrs := schedule.AsValueMap()
	require.Len(t, attrs, 2, "schedule must contain only writable settings")
	require.Equal(t, "hourly", attrs["frequency"].AsString())
	require.Equal(t, "15", attrs["minute"].AsString())
	schemas, diags := body.Attributes["schemas"].Expr.Value(ctx)
	require.False(t, diags.HasErrors())
	require.Equal(t, 1, schemas.LengthInt(), "disabled unselected tables must not be exported")
	schema := schemas.Index(cty.NumberIntVal(0)).AsValueMap()
	require.Equal(t, "secret", schema["excluded_fields"].Index(cty.NumberIntVal(0)).AsString())
	require.True(t, schema["history_enabled"].True())
	require.NotContains(t, schema, "destination_history_output_name", "computed history names are not configurable")
	require.NotContains(t, schema, "fields", "exclusions must not freeze the discovered column list")
}

func TestBulkSyncExportPreservesSchemaOverrides(t *testing.T) {
	b := NewBulkSyncs(newTestClient(t, map[string]string{
		"/api/bulk/syncs/sync-1/schemas": `{"data":[{"id":"contacts","enabled":true}]}`,
		"/api/bulk/syncs/sync-1/schemas/contacts": `{"data":{
			"id":"contacts","enabled":true,"excluded_fields":["secret"],"history_enabled":false,
			"partition_key":"id","tracking_field":"updated_at","user_output_name":"renamed_contacts",
			"disable_data_cutoff":true,"data_cutoff_timestamp":"2026-01-01T00:00:00Z",
			"fields":[{"id":"id","enabled":true},{"id":"email","enabled":true,"obfuscated":true,"user_output_name":"masked_email","output_name":"masked_email"},{"id":"secret","enabled":false}],
			"filters":[{"field_id":"status","function":"StringOneOf","value":["active","pending"]},{"field_id":"deleted_at","function":"IsNull"}]
		}}`,
	}))
	b.Resources["test"] = &polytomic.BulkSyncResponse{
		ID: pointer.ToString("sync-1"), Name: pointer.ToString("Test"), Active: pointer.ToBool(false), Mode: pointer.To(polytomic.BulkSyncTargetModeReplicate),
		SourceConnectionID: pointer.ToString("source"), DestinationConnectionID: pointer.ToString("destination"),
		DefaultSchedule: &polytomic.BulkSyncDefaultScheduleResponse{Frequency: polytomic.ScheduleFrequencyManual},
	}
	var output bytes.Buffer
	require.NoError(t, b.GenerateTerraformFiles(t.Context(), &output, nil))
	file, diags := hclsyntax.ParseConfig(output.Bytes(), "bulk_syncs.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors(), "%s", diags)
	value, diags := file.Body.(*hclsyntax.Body).Blocks[0].Body.Attributes["schemas"].Expr.Value(nil)
	require.False(t, diags.HasErrors(), "%s", diags)
	schema := value.Index(cty.NumberIntVal(0)).AsValueMap()
	require.Equal(t, "renamed_contacts", schema["user_output_name"].AsString())
	require.Equal(t, "id", schema["partition_key"].AsString())
	require.Equal(t, "updated_at", schema["tracking_field"].AsString())
	require.True(t, schema["disable_data_cutoff"].True())
	require.False(t, schema["history_enabled"].True())
	require.Equal(t, "2026-01-01T00:00:00Z", schema["data_cutoff_timestamp"].AsString())
	fields := schema["fields"]
	require.Equal(t, 3, fields.LengthInt(), "explicit overrides must match the imported field set")
	field := fields.Index(cty.NumberIntVal(1)).AsValueMap()
	require.True(t, field["obfuscate"].True())
	require.Equal(t, "masked_email", field["user_output_name"].AsString())
	require.NotContains(t, field, "output_name")
	filters := schema["filters"]
	require.Equal(t, 2, filters.LengthInt())
	require.JSONEq(t, `["active","pending"]`, filters.Index(cty.NumberIntVal(0)).AsValueMap()["value"].AsString())
	require.NotContains(t, filters.Index(cty.NumberIntVal(1)).AsValueMap(), "value")
}
