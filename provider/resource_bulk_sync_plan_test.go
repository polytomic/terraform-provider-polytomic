package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
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
		{name: "imported exclusions without a field snapshot", editState: func(s map[string]any) {
			s["schemas"].([]any)[0].(map[string]any)["excluded_fields"] = []any{"secret"}
		}, edit: func(c map[string]any) {
			delete(c["schemas"].([]any)[0].(map[string]any), "fields")
		}},
		{name: "clear exclusions", changed: true, editState: func(s map[string]any) {
			s["schemas"].([]any)[0].(map[string]any)["excluded_fields"] = []any{"secret"}
		}, edit: func(c map[string]any) {
			c["schemas"].([]any)[0].(map[string]any)["excluded_fields"] = []any{}
		}},
		{name: "omitted exclusions keep current list", editState: func(s map[string]any) {
			s["schemas"].([]any)[0].(map[string]any)["excluded_fields"] = []any{"secret"}
		}, edit: func(c map[string]any) {
			delete(c["schemas"].([]any)[0].(map[string]any), "excluded_fields")
		}},
		{name: "disable history", changed: true, editState: func(s map[string]any) {
			s["schemas"].([]any)[0].(map[string]any)["history_enabled"] = true
		}, edit: func(c map[string]any) {
			c["schemas"].([]any)[0].(map[string]any)["history_enabled"] = false
		}},
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

// Exercise real framework planning, SDK JSON requests, refresh, and replanning.
// An exclusion list is configuration; the changing discovered fields are not.
func TestBulkSyncExclusionsLifecycle(t *testing.T) {
	var mu sync.Mutex
	var saved map[string]any
	var excluded []any
	var historyEnabled bool
	discovered := []any{map[string]any{"id": "email", "enabled": true}, map[string]any{"id": "secret", "enabled": false}}
	writes := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/bulk/syncs", "/api/bulk/syncs/sync-1":
			if r.Method != http.MethodGet {
				var input map[string]any
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				selection := input["schemas"].([]any)[0].(map[string]any)
				if writes < 2 {
					if fields, ok := selection["fields"]; ok {
						t.Errorf("unconfigured discovered fields sent back to API: %v", fields)
					}
				} else {
					fields, ok := selection["fields"].([]any)
					if !ok || len(fields) != 1 {
						t.Errorf("explicit field selection was not sent: %v", fields)
					} else {
						field := fields[0].(map[string]any)
						if field["id"] != "email" || field["enabled"] != true || field["obfuscate"] != true {
							t.Errorf("explicit field override was not sent: %v", field)
						}
					}
				}
				if value, ok := selection["excluded_fields"]; ok {
					excluded = value.([]any)
				}
				if value, ok := selection["history_enabled"].(bool); ok {
					historyEnabled = value
				}
				if writes == 0 && len(excluded) != 1 {
					t.Error("create must send the exclusion")
				}
				if writes == 1 {
					if value, ok := selection["excluded_fields"]; !ok || len(value.([]any)) != 0 {
						t.Error("update must explicitly send [] to clear exclusions")
					}
				}
				writes++
				saved = input
				saved["id"] = "sync-1"
				saved["created_by"] = map[string]any{"id": "user-1", "name": "User", "type": "user"}
				saved["updated_by"] = saved["created_by"]
			}
			json.NewEncoder(w).Encode(map[string]any{"data": saved})
		case "/api/bulk/syncs/sync-1/schemas":
			fmt.Fprint(w, `{"data":[{"id":"contacts","enabled":true}]}`)
		case "/api/bulk/syncs/sync-1/schemas/contacts":
			json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{
				"id": "contacts", "enabled": true, "fields": discovered, "excluded_fields": excluded,
				"history_enabled": historyEnabled, "destination_history_output_name": "contacts_history",
			}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer api.Close()
	server := providerserver.NewProtocol6(New("test")())()
	schemas, err := server.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	dynamic := func(typ tftypes.Type, data any) *tfprotov6.DynamicValue {
		encoded, err := json.Marshal(data)
		require.NoError(t, err)
		value, err := tftypes.ValueFromJSON(encoded, typ)
		require.NoError(t, err)
		result, err := tfprotov6.NewDynamicValue(typ, value)
		require.NoError(t, err)
		return &result
	}
	configured, err := server.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{
		Config: dynamic(schemas.Provider.ValueType(), map[string]any{"api_key": "test", "deployment_url": api.URL}),
	})
	require.NoError(t, err)
	require.Empty(t, configured.Diagnostics)
	typ := schemas.ResourceSchemas["polytomic_bulk_sync"].ValueType()
	selection := map[string]any{"id": "contacts", "enabled": true, "excluded_fields": []string{"secret"}, "history_enabled": true}
	config := map[string]any{
		"name": "Exclusions", "active": false, "mode": "replicate",
		"automatically_add_new_fields": "all", "automatically_add_new_objects": "none",
		"source": map[string]any{"connection_id": "source"}, "destination": map[string]any{"connection_id": "destination"},
		"schedule": map[string]any{"frequency": "manual"}, "schemas": []any{selection},
	}
	state := dynamic(typ, nil)
	plan := func() *tfprotov6.PlanResourceChangeResponse {
		result, err := server.PlanResourceChange(t.Context(), &tfprotov6.PlanResourceChangeRequest{
			TypeName: "polytomic_bulk_sync", PriorState: state, Config: dynamic(typ, config), ProposedNewState: dynamic(typ, config),
		})
		require.NoError(t, err)
		require.Empty(t, result.Diagnostics)
		return result
	}
	apply := func() {
		planned := plan()
		result, err := server.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
			TypeName: "polytomic_bulk_sync", PriorState: state, Config: dynamic(typ, config), PlannedState: planned.PlannedState,
		})
		require.NoError(t, err)
		for _, diagnostic := range result.Diagnostics {
			t.Logf("%s: %s", diagnostic.Summary, diagnostic.Detail)
		}
		require.Empty(t, result.Diagnostics)
		state = result.NewState
		value, err := state.Unmarshal(typ)
		require.NoError(t, err)
		require.True(t, value.IsFullyKnown(), "applied state must not contain unknowns")
		var attrs map[string]tftypes.Value
		require.NoError(t, value.As(&attrs))
		var selections []tftypes.Value
		require.NoError(t, attrs["schemas"].As(&selections))
		var schema map[string]tftypes.Value
		require.NoError(t, selections[0].As(&schema))
		var history bool
		require.NoError(t, schema["history_enabled"].As(&history))
		require.Equal(t, selection["history_enabled"], history)
		var historyName string
		require.NoError(t, schema["destination_history_output_name"].As(&historyName))
		require.Equal(t, "contacts_history", historyName)
	}
	apply()
	mu.Lock()
	discovered = append(discovered, map[string]any{"id": "new_column", "enabled": true})
	mu.Unlock()
	refreshed, err := server.ReadResource(t.Context(), &tfprotov6.ReadResourceRequest{TypeName: "polytomic_bulk_sync", CurrentState: state})
	require.NoError(t, err)
	require.Empty(t, refreshed.Diagnostics)
	state = refreshed.NewState
	planned := plan()
	want, err := state.Unmarshal(typ)
	require.NoError(t, err)
	got, err := planned.PlannedState.Unmarshal(typ)
	require.NoError(t, err)
	require.True(t, got.Equal(want), "newly discovered columns must not create a plan diff")
	selection["excluded_fields"] = []string{}
	apply()
	mu.Lock()
	require.Empty(t, excluded)
	require.Equal(t, 2, writes)
	excluded = []any{"secret"}
	mu.Unlock()
	refreshed, err = server.ReadResource(t.Context(), &tfprotov6.ReadResourceRequest{TypeName: "polytomic_bulk_sync", CurrentState: state})
	require.NoError(t, err)
	require.Empty(t, refreshed.Diagnostics)
	state = refreshed.NewState
	want, err = state.Unmarshal(typ)
	require.NoError(t, err)
	got, err = plan().PlannedState.Unmarshal(typ)
	require.NoError(t, err)
	require.False(t, got.Equal(want), "out-of-band exclusion edits must still produce a corrective plan")
	selection["fields"] = []any{map[string]any{"id": "email", "enabled": true, "obfuscate": true}}
	selection["history_enabled"] = false
	apply()
}
