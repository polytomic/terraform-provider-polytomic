package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/polytomic/polytomic-go/v25"
)

func checkModelSyncAPI(t *testing.T, check func(*polytomic.ModelSyncV5Response) error) resource.TestCheckFunc {
	t.Helper()
	return func(state *terraform.State) error {
		rs := state.RootModule().Resources["polytomic_sync.test"]
		if rs == nil {
			return fmt.Errorf("model sync missing from state")
		}
		result, err := testClient(t, "").ModelSync.Get(t.Context(), rs.Primary.ID)
		if err != nil {
			return err
		}
		return check(result.Data)
	}
}

func TestAccSyncResourceBooleanFlagsUpdate(t *testing.T) {
	name := "TestAccSyncFlagsUpdate-" + uuid.NewString()
	cfg := func(enabled bool) string {
		value := fmt.Sprint(enabled)
		return syncAdvancedTestConfig(t, syncAdvancedTestArgs{Name: name, APIKey: APIKey(), Mode: "replace", Active: "false", Fields: defaultAdvancedSyncFields, SyncAllRecords: value, OnlyEnrichUpdates: value, SkipInitialBackfill: value})
	}
	check := func(enabled bool) resource.TestCheckFunc {
		return checkModelSyncAPI(t, func(sync *polytomic.ModelSyncV5Response) error {
			for key, value := range map[string]*bool{"sync_all_records": sync.SyncAllRecords, "only_enrich_updates": sync.OnlyEnrichUpdates, "skip_initial_backfill": sync.SkipInitialBackfill} {
				if value == nil || *value != enabled {
					return fmt.Errorf("API %s = %v, want %t", key, value, enabled)
				}
			}
			return nil
		})
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: cfg(false), Check: check(false)}, {Config: cfg(true), Check: check(true)}, {Config: cfg(true), PlanOnly: true}, {Config: cfg(false), Check: check(false)}, {Config: cfg(false), PlanOnly: true},
	}})
}

func TestAccSyncResourceTargetFiltersLifecycle(t *testing.T) {
	target := os.Getenv("POLYTOMIC_SYNC_TEST_SALESFORCE_CONNECTION_ID")
	if target == "" {
		t.Skip("requires POLYTOMIC_SYNC_TEST_SALESFORCE_CONNECTION_ID")
	}
	name := "TestAccSyncTargetFilters-" + uuid.NewString()
	cfg := func(filters string) string {
		return syncAdvancedTestConfig(t, syncAdvancedTestArgs{
			Name: name, APIKey: APIKey(), Mode: "update", Active: "false", ModelQuery: identityModelQuery,
			Fields:   `[{source={field="name",model_id=polytomic_model.test.id},target="LastName"}]`,
			Identity: `{source={field="email",model_id=polytomic_model.test.id},target="Email",function="Equality"}`,
			Target:   fmt.Sprintf(`{connection_id=%q,object="Contact"}`, target), TargetFilters: filters,
		})
	}
	check := func(expected int, value string) resource.TestCheckFunc {
		return checkModelSyncAPI(t, func(sync *polytomic.ModelSyncV5Response) error {
			if pointer.GetBool(sync.Active) {
				return fmt.Errorf("fixture must remain inactive")
			}
			if len(sync.Filters) != expected {
				return fmt.Errorf("API filter count = %d, want %d", len(sync.Filters), expected)
			}
			if expected > 0 {
				f := sync.Filters[0]
				if pointer.Get(f.FieldType) != polytomic.FilterFieldReferenceTypeTarget || pointer.GetString(f.FieldID) != "Email" || f.Value != value {
					return fmt.Errorf("API target filter differs: field=%v, value=%v", f.FieldID, f.Value)
				}
			}
			return nil
		})
	}
	empty := cfg("[]")
	first := cfg(`[{field="Email",function="Equality",value=jsonencode("first@example.com")}]`)
	second := cfg(`[{field="Email",function="Equality",value=jsonencode("second@example.com")}]`)
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: first, Check: check(1, "first@example.com")}, {Config: first, PlanOnly: true},
		{Config: second, Check: check(1, "second@example.com")}, {Config: second, PlanOnly: true},
		{Config: empty, Check: check(0, "")}, {Config: empty, PlanOnly: true},
	}})
}

func TestAccSyncResourceScheduleUpdate(t *testing.T) {
	name := "TestAccSyncScheduleUpdate-" + uuid.NewString()
	cfg := func(schedule string) string {
		return syncAdvancedTestConfig(t, syncAdvancedTestArgs{Name: name, APIKey: APIKey(), Mode: "replace", Active: "false", Fields: defaultAdvancedSyncFields, Schedule: schedule})
	}
	check := func(frequency string) resource.TestCheckFunc {
		return checkModelSyncAPI(t, func(sync *polytomic.ModelSyncV5Response) error {
			if sync.Schedule == nil || string(pointer.Get(sync.Schedule.Frequency)) != frequency {
				return fmt.Errorf("API schedule frequency differs from %s", frequency)
			}
			if frequency == "daily" && (pointer.GetString(sync.Schedule.Hour) != "3" || pointer.GetString(sync.Schedule.Minute) != "15") {
				return fmt.Errorf("API schedule time differs from 03:15")
			}
			return nil
		})
	}
	manual := cfg(`{frequency="manual"}`)
	daily := cfg(`{frequency="daily",hour="3",minute="15"}`)
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: manual, Check: check("manual")}, {Config: daily, Check: check("daily")}, {Config: daily, PlanOnly: true}, {Config: manual, Check: check("manual")}, {Config: manual, PlanOnly: true},
	}})
}

func TestAccSyncResourceFieldMappingsUpdate(t *testing.T) {
	name := "TestAccSyncMappingsUpdate-" + uuid.NewString()
	email := `{source={field="email",model_id=polytomic_model.test.id},target="email"}`
	field := `{source={field="name",model_id=polytomic_model.test.id},target="name"}`
	cfg := func(fields, overrides string) string {
		return syncAdvancedTestConfig(t, syncAdvancedTestArgs{Name: name, APIKey: APIKey(), Mode: "replace", Active: "false", ModelQuery: identityModelQuery, Fields: fields, OverrideFields: overrides})
	}
	check := func(fields, overrides int) resource.TestCheckFunc {
		return checkModelSyncAPI(t, func(sync *polytomic.ModelSyncV5Response) error {
			if len(sync.Fields) != fields || len(sync.OverrideFields) != overrides {
				return fmt.Errorf("API mappings/overrides = %d/%d, want %d/%d", len(sync.Fields), len(sync.OverrideFields), fields, overrides)
			}
			if overrides > 0 && (sync.OverrideFields[0].Target != "name" || sync.OverrideFields[0].OverrideValue != "default") {
				return fmt.Errorf("API static override differs from default name")
			}
			return nil
		})
	}
	full := cfg("["+email+","+field+"]", "[]")
	reordered := cfg("["+field+","+email+"]", "[]")
	reduced := cfg("["+email+"]", "[]")
	static := cfg("["+email+"]", `[{target="name",override_value="default"}]`)
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: full, Check: check(2, 0)}, {Config: reordered, PlanOnly: true},
		{Config: reduced, Check: check(1, 0)}, {Config: reduced, PlanOnly: true},
		// The local backend rejects an empty unconditional override as a missing
		// source field. Retain this limitation as an explicit rejection test.
		{Config: cfg("["+email+"]", `[{target="name",override_value=""}]`), ExpectError: regexp.MustCompile("unknown model field id")},
		{Config: static, Check: check(1, 1)}, {Config: static, PlanOnly: true},
		{Config: full, Check: check(2, 0)}, {Config: full, PlanOnly: true},
	}})
}
