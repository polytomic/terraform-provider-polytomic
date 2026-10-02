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
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func TestBulkSyncExportPreservesConfiguration(t *testing.T) {
	b := NewBulkSyncs(newTestClient(t, map[string]string{
		"/api/bulk/syncs/sync-1/schemas": `{"data":[{"id":"contacts","enabled":true},{"id":"unselected","enabled":false}]}`,
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
}
