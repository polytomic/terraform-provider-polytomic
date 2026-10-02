package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// Exercise the framework's unknown marking and schema modifiers, rather than
// calling ModifyPlan directly. Set elements cannot be correlated by position.
func TestBulkSyncImportedPlan(t *testing.T) {
	server := providerserver.NewProtocol6(New("test")())()
	schemas, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	typ := schemas.ResourceSchemas["polytomic_bulk_sync"].ValueType()
	field := func(id string) map[string]any {
		return map[string]any{"id": id, "enabled": true, "obfuscate": false, "output_name": id, "user_output_name": nil}
	}
	state := map[string]any{
		"id": "sync-1", "name": "Imported", "active": false, "mode": "replicate",
		"automatically_add_new_fields": "none", "automatically_add_new_objects": "none",
		"source":      map[string]any{"connection_id": "source", "configuration": "{}"},
		"destination": map[string]any{"connection_id": "destination", "configuration": "{}"},
		"schedule":    map[string]any{"frequency": "manual"},
		"schemas":     []any{map[string]any{"id": "contacts", "enabled": true, "output_name": "contacts", "partition_key": "", "tracking_field": "", "disable_data_cutoff": false, "fields": []any{field("email"), field("name")}}},
		"created_at":  "2026-09-09T18:39:11Z", "updated_at": "2026-09-30T15:33:32Z",
		"created_by": map[string]any{"id": "user-1", "name": "User", "type": "user"},
		"updated_by": map[string]any{"id": "user-1", "name": "User", "type": "user"},
	}
	clone := func(v map[string]any) map[string]any {
		b, e := json.Marshal(v)
		require.NoError(t, e)
		var result map[string]any
		require.NoError(t, json.Unmarshal(b, &result))
		return result
	}
	dynamic := func(v map[string]any) *tfprotov6.DynamicValue {
		b, e := json.Marshal(v)
		require.NoError(t, e)
		value, e := tftypes.ValueFromJSON(b, typ)
		require.NoError(t, e)
		if v["name"] == "<unknown>" {
			var attrs map[string]tftypes.Value
			require.NoError(t, value.As(&attrs))
			attrs["name"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
			value = tftypes.NewValue(typ, attrs)
		}
		d, e := tfprotov6.NewDynamicValue(typ, value)
		require.NoError(t, e)
		return &d
	}
	for _, tc := range []struct {
		name      string
		edit      func(map[string]any)
		editState func(map[string]any)
		changed   bool
	}{
		{name: "generated configuration"},
		{name: "default still changes discovery", changed: true,
			editState: func(s map[string]any) { s["automatically_add_new_fields"] = "all" },
			edit:      func(c map[string]any) { delete(c, "automatically_add_new_fields") }},
		{name: "cutoff removal", changed: true,
			editState: func(s map[string]any) {
				s["schemas"].([]any)[0].(map[string]any)["data_cutoff_timestamp"] = "2026-01-01T00:00:00Z"
			},
			edit: func(c map[string]any) { delete(c["schemas"].([]any)[0].(map[string]any), "data_cutoff_timestamp") }},
		{name: "schemas left to server", edit: func(c map[string]any) { delete(c, "schemas") }},
		{name: "schema tracking field edit", changed: true, edit: func(c map[string]any) { c["schemas"].([]any)[0].(map[string]any)["tracking_field"] = "email" }},
		{name: "field override edit", changed: true, edit: func(c map[string]any) {
			c["schemas"].([]any)[0].(map[string]any)["fields"].([]any)[0].(map[string]any)["user_output_name"] = "renamed"
		}},
		{name: "removed schema", changed: true, edit: func(c map[string]any) { c["schemas"] = []any{} }},
		{name: "schedule edit", changed: true, edit: func(c map[string]any) { c["schedule"] = map[string]any{"frequency": "hourly"} }},
		{name: "connection edit", changed: true, edit: func(c map[string]any) { c["source"].(map[string]any)["configuration"] = `{"changed":true}` }},

		{name: "importer schemas only", edit: func(c map[string]any) { c["schemas"] = []any{map[string]any{"id": "contacts", "enabled": true}} }},
		{name: "field edit", changed: true, edit: func(c map[string]any) {
			c["schemas"].([]any)[0].(map[string]any)["fields"].([]any)[0].(map[string]any)["enabled"] = false
		}},
		{name: "name edit", changed: true, edit: func(c map[string]any) { c["name"] = "Changed" }},
		{name: "new schema", changed: true, edit: func(c map[string]any) {
			c["schemas"] = append(c["schemas"].([]any), map[string]any{"id": "new", "enabled": true})
		}},
		{name: "removed field", changed: true, edit: func(c map[string]any) {
			c["schemas"].([]any)[0].(map[string]any)["fields"] = []any{map[string]any{"id": "email", "enabled": true, "obfuscate": false}}
		}},
		{name: "unknown configured value", changed: true, edit: func(c map[string]any) { c["name"] = "<unknown>" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			priorState := clone(state)
			if tc.editState != nil {
				tc.editState(priorState)
			}
			config := clone(priorState)
			for _, k := range []string{"id", "created_at", "updated_at", "created_by", "updated_by"} {
				delete(config, k)
			}
			s := config["schemas"].([]any)[0].(map[string]any)
			delete(s, "partition_key")
			delete(s, "tracking_field")
			fields := s["fields"].([]any)
			// Reverse order and omit computed output names, as generated HCL does.
			fields[0], fields[1] = fields[1], fields[0]
			for _, f := range fields {
				delete(f.(map[string]any), "output_name")
				delete(f.(map[string]any), "user_output_name")
			}
			if tc.edit != nil {
				tc.edit(config)
			}
			proposed := clone(config)
			for _, k := range []string{"id", "created_at", "updated_at", "created_by", "updated_by"} {
				proposed[k] = priorState[k]
			}
			result, e := server.PlanResourceChange(t.Context(), &tfprotov6.PlanResourceChangeRequest{TypeName: "polytomic_bulk_sync", PriorState: dynamic(priorState), Config: dynamic(config), ProposedNewState: dynamic(proposed)})
			require.NoError(t, e)
			require.Empty(t, result.Diagnostics)
			planned, e := result.PlannedState.Unmarshal(typ)
			require.NoError(t, e)
			prior, e := dynamic(priorState).Unmarshal(typ)
			require.NoError(t, e)
			require.Equal(t, !tc.changed, planned.Equal(prior))
			if tc.changed {
				var attrs map[string]tftypes.Value
				require.NoError(t, planned.As(&attrs))
				require.False(t, attrs["updated_at"].IsKnown(), "real changes must allow audit fields to change")
				require.False(t, attrs["updated_by"].IsKnown())
			}
		})
	}
}
