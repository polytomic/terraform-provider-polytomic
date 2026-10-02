package importer

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/polytomic/polytomic-go/v25"
	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func TestSyncExportPreservesJSONValuesAndFlags(t *testing.T) {
	for name, value := range map[string]any{"empty_string": "", "null": nil, "empty_array": []any{}, "mixed": []any{"", nil, false, []any{1.0}, map[string]any{"empty": ""}}} {
		t.Run(name, func(t *testing.T) {
			exporter := NewSyncs(nil)
			configuration := map[string]any{"empty": "", "value": value}
			exporter.Resources["test"] = &polytomic.ModelSyncV5Response{
				Name: pointer.ToString("Test"), Mode: pointer.To(polytomic.ModelsyncSyncTargetModeUpdate),
				Schedule: &polytomic.Schedule{Frequency: pointer.To(polytomic.ScheduleFrequencyManual)},
				Target:   &polytomic.ModelSyncV5Target{ConnectionID: "destination", Object: pointer.ToString("contacts"), Configuration: configuration},
				Filters: []*polytomic.Filter{
					{Field: &polytomic.Source{ModelID: "model", Field: "email"}, Function: polytomic.FilterFunctionEquality, Value: value},
					{FieldType: pointer.To(polytomic.FilterFieldReferenceTypeTarget), FieldID: pointer.ToString("Email"), Function: polytomic.FilterFunctionEquality, Value: value},
				},
				Overrides:      []*polytomic.Override{{Field: &polytomic.Source{ModelID: "model", Field: "email"}, Function: pointer.To(polytomic.FilterFunctionEquality), Value: value, Override: value}},
				Fields:         []*polytomic.SyncField{{Source: &polytomic.Source{ModelID: "model", Field: "email"}, Target: "email", OverrideValue: pointer.ToString("")}},
				OverrideFields: []*polytomic.OverrideField{{Target: "name", OverrideValue: ""}},
				SyncAllRecords: pointer.ToBool(true), OnlyEnrichUpdates: pointer.ToBool(true), SkipInitialBackfill: pointer.ToBool(true),
			}
			var output bytes.Buffer
			require.NoError(t, exporter.GenerateTerraformFiles(t.Context(), &output, nil))
			file, diags := hclsyntax.ParseConfig(output.Bytes(), "syncs.tf", hcl.InitialPos)
			require.False(t, diags.HasErrors(), "%s", diags)
			body := file.Body.(*hclsyntax.Body).Blocks[0].Body
			ctx := &hcl.EvalContext{Functions: map[string]function.Function{"jsonencode": stdlib.JSONEncodeFunc}}
			for _, key := range []string{"sync_all_records", "only_enrich_updates", "skip_initial_backfill"} {
				require.Contains(t, body.Attributes, key)
				actual, diags := body.Attributes[key].Expr.Value(ctx)
				require.False(t, diags.HasErrors())
				require.True(t, actual.True())
			}
			for _, key := range []string{"fields", "override_fields"} {
				actual, diags := body.Attributes[key].Expr.Value(ctx)
				require.False(t, diags.HasErrors())
				require.Equal(t, "", actual.AsValueSlice()[0].AsValueMap()["override_value"].AsString())
			}
			expected, err := json.Marshal(value)
			require.NoError(t, err)
			for _, key := range []string{"filters", "target_filters", "overrides"} {
				actual, diags := body.Attributes[key].Expr.Value(ctx)
				require.False(t, diags.HasErrors(), "%s", diags)
				first := actual.AsValueSlice()[0].AsValueMap()
				if value == nil {
					require.NotContains(t, first, "value")
				} else {
					require.JSONEq(t, string(expected), first["value"].AsString())
				}
				if key == "overrides" {
					require.JSONEq(t, string(expected), first["override"].AsString())
				}
			}
			target, diags := body.Attributes["target"].Expr.Value(ctx)
			require.False(t, diags.HasErrors())
			expected, err = json.Marshal(configuration)
			require.NoError(t, err)
			require.JSONEq(t, string(expected), target.AsValueMap()["configuration"].AsString())
		})
	}
}
