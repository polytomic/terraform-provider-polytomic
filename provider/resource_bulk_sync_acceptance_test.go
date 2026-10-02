package provider

import (
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/polytomic/polytomic-go/v25"
)

// Check the API as well as Terraform state: merging the plan into state can
// otherwise make a request that omitted a setting appear to succeed.
func checkBulkSyncAPI(t *testing.T, check func(*polytomic.BulkSyncResponse) error) resource.TestCheckFunc {
	t.Helper()
	return func(state *terraform.State) error {
		rs := state.RootModule().Resources["polytomic_bulk_sync.test"]
		if rs == nil {
			return fmt.Errorf("bulk sync missing from state")
		}
		result, err := testClient(t, "").BulkSync.Get(t.Context(), rs.Primary.ID, &polytomic.BulkSyncGetRequest{})
		if err != nil {
			return err
		}
		return check(result.Data)
	}
}

func TestAccBulkSyncResourceScheduleUpdate(t *testing.T) {
	conns := getBulkSyncConnections(t)
	name := "TestAccBulkSyncSchedule-" + uuid.NewString()
	config := func(frequency, extra string) string {
		return fmt.Sprintf(`resource "polytomic_bulk_sync" "test" {
   name = %q
   active = false
   mode = "replicate"
   source = { connection_id = %q }
   destination = { connection_id = %q, configuration = jsonencode({schema = "public"}) }
   schemas = [{ id = "polytomic.sync_test_source", enabled = true }]
   schedule = { frequency = %q %s }
  }`, name, conns.SourceID, conns.DestID, frequency, extra)
	}
	var scheduleID string
	check := func(frequency string) resource.TestCheckFunc {
		return checkBulkSyncAPI(t, func(sync *polytomic.BulkSyncResponse) error {
			if sync.DefaultSchedule == nil {
				return fmt.Errorf("default schedule missing")
			}
			id := pointer.GetString(sync.DefaultSchedule.ID)
			if id == "" {
				return fmt.Errorf("default schedule ID missing")
			}
			if scheduleID == "" {
				scheduleID = id
			} else if scheduleID != id {
				return fmt.Errorf("schedule identity changed: %s -> %s", scheduleID, id)
			}
			if string(sync.DefaultSchedule.Frequency) != frequency {
				return fmt.Errorf("API frequency = %s, want %s", sync.DefaultSchedule.Frequency, frequency)
			}
			return nil
		})
	}
	manual := config("manual", "")
	daily := config("daily", `, hour = "3", minute = "15"`)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: manual, Check: check("manual")},
			{Config: daily, Check: check("daily")},
			{Config: daily, PlanOnly: true},
			{Config: manual, Check: check("manual")},
			{Config: manual, PlanOnly: true},
		},
	})
}

func TestAccBulkSyncResourceNestedUpdates(t *testing.T) {
	conns := getBulkSyncConnections(t)
	name := "TestAccBulkSyncNested-" + uuid.NewString()
	config := func(enabled, obfuscate bool, orderReversed bool) string {
		source := fmt.Sprintf(`{ id = "polytomic.sync_test_source", enabled = true,
   fields = [{id = "email", enabled = true, obfuscate = %t}, {id = "name", enabled = %t, obfuscate = false}]
  }`, obfuscate, enabled)
		other := `{id = "polytomic.sync_test_other", enabled = true}`
		schemas := source + "," + other
		if orderReversed {
			schemas = other + "," + source
		}
		return bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{Name: name, SourceConnectionID: conns.SourceID, DestConnectionID: conns.DestID, Mode: "replicate", Active: "false", Schemas: "[" + schemas + "]"})
	}
	check := func(enabled, obfuscate bool) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			sync := state.RootModule().Resources["polytomic_bulk_sync.test"]
			schema, err := testClient(t, "").BulkSync.Schemas.Get(t.Context(), sync.Primary.ID, "polytomic.sync_test_source")
			if err != nil {
				return err
			}
			found := map[string]bool{}
			for _, field := range schema.Data.Fields {
				id := pointer.GetString(field.ID)
				if id == "email" {
					found[id] = true
					if !pointer.GetBool(field.Enabled) || pointer.GetBool(field.Obfuscated) != obfuscate {
						return fmt.Errorf("API email configuration not updated")
					}
				}
				if id == "name" {
					found[id] = true
					if pointer.GetBool(field.Enabled) != enabled {
						return fmt.Errorf("API name enabled = %v, want %v", field.Enabled, enabled)
					}
				}
			}
			if !found["email"] || !found["name"] {
				return fmt.Errorf("API missing configured fields")
			}
			return nil
		}
	}
	initial := config(true, false, false)
	updated := config(false, true, false)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: initial, Check: check(true, false)},
			{Config: config(true, false, true), PlanOnly: true},
			{Config: updated, Check: check(false, true)},
			{Config: updated, PlanOnly: true},
			{Config: initial, Check: check(true, false)},
			{Config: initial, PlanOnly: true},
		},
	})
}

func TestAccBulkSyncResourceOptionsUpdate(t *testing.T) {
	conns := getBulkSyncConnections(t)
	name := "TestAccBulkSyncOptionsUpdate-" + uuid.NewString()
	config := func(concurrency, resync int, timestamps bool, normalize string) string {
		return bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{
			Name: name, SourceConnectionID: conns.SourceID, DestConnectionID: conns.DestID, Mode: "replicate", Active: "false",
			Schemas:          `[{id = "polytomic.sync_test_source", enabled = true}]`,
			ConcurrencyLimit: fmt.Sprint(concurrency), ResyncConcurrencyLimit: fmt.Sprint(resync),
			DisableRecordTimestamps: fmt.Sprint(timestamps), NormalizeNames: normalize,
		})
	}
	check := func(concurrency, resync int, timestamps bool, normalize string) resource.TestCheckFunc {
		return checkBulkSyncAPI(t, func(sync *polytomic.BulkSyncResponse) error {
			if pointer.GetInt(sync.ConcurrencyLimit) != concurrency || pointer.GetInt(sync.ResyncConcurrencyLimit) != resync {
				return fmt.Errorf("API concurrency limits do not match configuration")
			}
			if pointer.GetBool(sync.DisableRecordTimestamps) != timestamps {
				return fmt.Errorf("API disable_record_timestamps does not match configuration")
			}
			if string(pointer.Get(sync.NormalizeNames)) != normalize {
				return fmt.Errorf("API normalize_names does not match configuration")
			}
			return nil
		})
	}
	initial := config(2, 3, false, "enabled")
	updated := config(4, 5, true, "disabled")
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: initial, Check: check(2, 3, false, "enabled")},
		{Config: updated, Check: check(4, 5, true, "disabled")},
		{Config: updated, PlanOnly: true},
		{Config: initial, Check: check(2, 3, false, "enabled")},
		{Config: initial, PlanOnly: true},
	}})
}

func TestAccBulkSyncResourceFiltersUpdate(t *testing.T) {
	conns := getBulkSyncConnections(t)
	name := "TestAccBulkSyncFiltersUpdate-" + uuid.NewString()
	config := func(value string) string {
		filters := "[]"
		if value != "" {
			filters = fmt.Sprintf(`[{field_id = "created_at", function = "OnOrAfter", value = jsonencode(%q)}]`, value)
		}
		return bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{Name: name, SourceConnectionID: conns.SourceID, DestConnectionID: conns.DestID, Mode: "replicate", Active: "false", Schemas: fmt.Sprintf(`[{id = "polytomic.sync_test_source", enabled = true, filters = %s}]`, filters)})
	}
	check := func(value string) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			rs := state.RootModule().Resources["polytomic_bulk_sync.test"]
			response, err := testClient(t, "").BulkSync.Schemas.Get(t.Context(), rs.Primary.ID, "polytomic.sync_test_source")
			if err != nil {
				return err
			}
			filters := response.Data.Filters
			if value == "" {
				if len(filters) != 0 {
					return fmt.Errorf("API filters were not cleared")
				}
				return nil
			}
			if len(filters) != 1 || pointer.GetString(filters[0].FieldID) != "created_at" || string(filters[0].Function) != "OnOrAfter" || filters[0].Value != value {
				return fmt.Errorf("API filter does not match configuration")
			}
			return nil
		}
	}
	first, second, cleared := config("2024-01-01"), config("2025-01-01"), config("")
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: first, Check: check("2024-01-01")},
		{Config: second, Check: check("2025-01-01")},
		{Config: second, PlanOnly: true},
		{Config: cleared, Check: check("")},
		{Config: cleared, PlanOnly: true},
	}})
}

func TestAccBulkSyncResourceOutputNameOverrides(t *testing.T) {
	conns := getBulkSyncConnections(t)
	name := "TestAccBulkSyncNames-" + uuid.NewString()
	config := func(table, column string) string {
		return bulkSyncAdvancedTestConfig(t, bulkSyncAdvancedTestArgs{Name: name, SourceConnectionID: conns.SourceID, DestConnectionID: conns.DestID, Mode: "replicate", Active: "false", Schemas: fmt.Sprintf(`[{id = "polytomic.sync_test_source", enabled = true, user_output_name = %q,
    fields = [{id = "email", enabled = true}, {id = "name", enabled = true, user_output_name = %q}]
  }]`, table, column)})
	}
	check := func(table, column string) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			rs := state.RootModule().Resources["polytomic_bulk_sync.test"]
			response, err := testClient(t, "").BulkSync.Schemas.Get(t.Context(), rs.Primary.ID, "polytomic.sync_test_source")
			if err != nil {
				return err
			}
			if pointer.GetString(response.Data.UserOutputName) != table {
				return fmt.Errorf("API table override = %q, want %q", pointer.GetString(response.Data.UserOutputName), table)
			}
			for _, field := range response.Data.Fields {
				if pointer.GetString(field.ID) == "name" {
					if pointer.GetString(field.UserOutputName) != column {
						return fmt.Errorf("API column override = %q, want %q", pointer.GetString(field.UserOutputName), column)
					}
					return nil
				}
			}
			return fmt.Errorf("API name field missing")
		}
	}
	initial, updated, cleared := config("tf_contacts", "tf_name"), config("tf_contacts_updated", "tf_name_updated"), config("", "")
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: initial, Check: check("tf_contacts", "tf_name")},
		{Config: updated, Check: check("tf_contacts_updated", "tf_name_updated")},
		{Config: updated, PlanOnly: true},
		{Config: cleared, Check: check("", "")},
		{Config: cleared, PlanOnly: true},
	}})
}
