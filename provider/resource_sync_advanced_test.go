package provider

import (
	"fmt"
	"strings"
	"testing"
	texttemplate "text/template"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

// syncAdvancedTestArgs holds template arguments for advanced sync acceptance tests.
type syncAdvancedTestArgs struct {
	Name                 string
	APIKey               bool
	Postgres             postgresTestConfig
	Mode                 string
	Active               string // "true" or "false"
	ModelQuery           string // SQL query for model (empty = default)
	Fields               string // Raw HCL for fields
	OverrideFields       string // Raw HCL for override_fields (empty = omit)
	Target               string // Raw HCL for target (empty = use default)
	Schedule             string // Raw HCL for schedule (empty = use default manual)
	Identity             string // Raw HCL for identity (empty = omit)
	TargetFilters        string // Raw HCL for target_filters (empty = omit)
	SyncAllRecords       string // "true" or "false" (empty = omit)
	OnlyEnrichUpdates    string
	SkipInitialBackfill  string
	EncryptionPassphrase string
}

const defaultAdvancedSyncFields = `[
    {
      source = {
        field    = "email"
        model_id = polytomic_model.test.id
      }
      target = "email"
    }
  ]`

// identitySyncFields uses the "name" column (from the multi-column model query)
// to avoid the server-side dedup bug when identity and a field both target "email".
const identitySyncFields = `[
    {
      source = {
        field    = "name"
        model_id = polytomic_model.test.id
      }
      target = "name"
    }
  ]`

const identityModelQuery = "SELECT email, email as name FROM polytomic.sync_test_source"

func syncAdvancedTestConfig(t *testing.T, args syncAdvancedTestArgs) string {
	t.Helper()
	args.Postgres = testPostgresConfig(t)
	tmpl := texttemplate.Must(texttemplate.New("sync-advanced").Parse(syncAdvancedTestTemplate))
	var buf strings.Builder
	require.NoError(t, tmpl.Execute(&buf, args))
	return buf.String()
}

const syncAdvancedTestTemplate = `
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
{{- if .ModelQuery}}
    "query" = "{{.ModelQuery}}"
{{- else}}
    "query" = "SELECT email FROM polytomic.sync_test_source"
{{- end}}
  })
  connection_id = polytomic_postgresql_connection.test.id
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}

resource "polytomic_sync" "test" {
  name   = "{{.Name}}"
  mode   = "{{.Mode}}"
  active = {{.Active}}
{{- if .Schedule}}
  schedule = {{.Schedule}}
{{- else}}
  schedule = {
    frequency = "manual"
  }
{{- end}}
  fields = {{.Fields}}
{{- if .OverrideFields}}
  override_fields = {{.OverrideFields}}
{{- end}}
{{- if .Target}}
  target = {{.Target}}
{{- else}}
  target = {
    connection_id = polytomic_postgresql_connection.test.id
    object        = "polytomic.sync_test_target"
  }
{{- end}}
{{- if .Identity}}
  identity = {{.Identity}}
{{- end}}
{{- if .TargetFilters}}
  target_filters = {{.TargetFilters}}
{{- end}}
{{- if .SyncAllRecords}}
  sync_all_records = {{.SyncAllRecords}}
{{- end}}
{{- if .OnlyEnrichUpdates}}
  only_enrich_updates = {{.OnlyEnrichUpdates}}
{{- end}}
{{- if .SkipInitialBackfill}}
  skip_initial_backfill = {{.SkipInitialBackfill}}
{{- end}}
{{- if .EncryptionPassphrase}}
  encryption_passphrase = "{{.EncryptionPassphrase}}"
{{- end}}
{{if not .APIKey}}
  organization = polytomic_organization.test.id
{{end}}
}
`

// ---------------------------------------------------------------------------
// API verification helpers
// ---------------------------------------------------------------------------

// testAccSyncHasIdentity verifies the sync has an identity configured via the API.
func testAccSyncHasIdentity(t *testing.T, apiKey bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var orgID string
		if !apiKey {
			org, ok := s.RootModule().Resources["polytomic_organization.test"]
			if !ok {
				return fmt.Errorf("not found: polytomic_organization.test")
			}
			orgID = org.Primary.ID
		}
		r, ok := s.RootModule().Resources["polytomic_sync.test"]
		if !ok {
			return fmt.Errorf("not found: polytomic_sync.test")
		}
		client := testClient(t, orgID)
		sync, err := client.ModelSync.Get(t.Context(), r.Primary.ID)
		if err != nil {
			return err
		}
		if sync.Data.Identity == nil {
			return fmt.Errorf("expected identity to be set, got nil")
		}
		return nil
	}
}

// testAccSyncMode verifies the sync mode via the API.
func testAccSyncMode(t *testing.T, apiKey bool, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var orgID string
		if !apiKey {
			org, ok := s.RootModule().Resources["polytomic_organization.test"]
			if !ok {
				return fmt.Errorf("not found: polytomic_organization.test")
			}
			orgID = org.Primary.ID
		}
		r, ok := s.RootModule().Resources["polytomic_sync.test"]
		if !ok {
			return fmt.Errorf("not found: polytomic_sync.test")
		}
		client := testClient(t, orgID)
		sync, err := client.ModelSync.Get(t.Context(), r.Primary.ID)
		if err != nil {
			return err
		}
		actual := string(pointer.Get(sync.Data.Mode))
		if actual != expected {
			return fmt.Errorf("expected mode %q, got %q", expected, actual)
		}
		return nil
	}
}

// testAccSyncOverrideFieldCount verifies the number of override fields via the
// API. The server merges override fields into the regular fields list, so we
// count fields that have an override_value and no real source.
func testAccSyncOverrideFieldCount(t *testing.T, apiKey bool, expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var orgID string
		if !apiKey {
			org, ok := s.RootModule().Resources["polytomic_organization.test"]
			if !ok {
				return fmt.Errorf("not found: polytomic_organization.test")
			}
			orgID = org.Primary.ID
		}
		r, ok := s.RootModule().Resources["polytomic_sync.test"]
		if !ok {
			return fmt.Errorf("not found: polytomic_sync.test")
		}
		client := testClient(t, orgID)
		sync, err := client.ModelSync.Get(t.Context(), r.Primary.ID)
		if err != nil {
			return err
		}
		count := 0
		for _, f := range sync.Data.Fields {
			if f.OverrideValue != nil &&
				(f.Source == nil || f.Source.ModelId == "" || f.Source.ModelId == "00000000-0000-0000-0000-000000000000") {
				count++
			}
		}
		if count != expected {
			return fmt.Errorf("expected %d override fields, got %d", expected, count)
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// Test: updateOrCreate mode with identity
// ---------------------------------------------------------------------------

func TestAccSyncResourceWithIdentity(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncIdentity-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:       name,
					APIKey:     apiKey,
					Mode:       "updateOrCreate",
					Active:     "false",
					ModelQuery: identityModelQuery,
					Fields:     identitySyncFields,
					Identity: `{
      source = {
        model_id = polytomic_model.test.id
        field    = "email"
      }
      target   = "email"
      function = "Equality"
    }`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("mode"),
						knownvalue.StringExact("updateOrCreate"),
					),
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("identity").AtMapKey("target"),
						knownvalue.StringExact("email"),
					),
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("identity").AtMapKey("function"),
						knownvalue.StringExact("Equality"),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
					testAccSyncHasIdentity(t, apiKey),
					testAccSyncMode(t, apiKey, "updateOrCreate"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: override_fields (unconditional static values)
// ---------------------------------------------------------------------------

func TestAccSyncResourceOverrideFields(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncOvFields-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with one override field.
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:   name,
					APIKey: apiKey,
					Mode:   "replace",
					Active: "false",
					Fields: defaultAdvancedSyncFields,
					OverrideFields: `[
    {
      target         = "name"
      override_value = "default"
    }
  ]`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("override_fields"),
						knownvalue.SetSizeExact(1),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
					testAccSyncOverrideFieldCount(t, apiKey, 1),
				),
			},
			// Step 2: Remove override fields.
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:   name,
					APIKey: apiKey,
					Mode:   "replace",
					Active: "false",
					Fields: defaultAdvancedSyncFields,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccSyncOverrideFieldCount(t, apiKey, 0),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Boolean flags (sync_all_records, skip_initial_backfill)
// ---------------------------------------------------------------------------

func TestAccSyncResourceBooleanFlags(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncFlags-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:                name,
					APIKey:              apiKey,
					Mode:                "replace",
					Active:              "false",
					Fields:              defaultAdvancedSyncFields,
					SyncAllRecords:      "true",
					SkipInitialBackfill: "true",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("sync_all_records"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("skip_initial_backfill"),
						knownvalue.Bool(true),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Field with override_value (static value, no source mapping)
// ---------------------------------------------------------------------------

// TestAccSyncResourceFieldOverrideValue tests a field with a source AND an
// override_value. When override_value is set, the source is ignored and the
// static value is used instead.
func TestAccSyncResourceFieldOverrideValue(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncFieldOv-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:   name,
					APIKey: apiKey,
					Mode:   "replace",
					Active: "false",
					Fields: `[
    {
      source = {
        field    = "email"
        model_id = polytomic_model.test.id
      }
      target = "email"
    },
    {
      source = {
        field    = "email"
        model_id = polytomic_model.test.id
      }
      target         = "name"
      override_value = "synced"
    }
  ]`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("fields"),
						knownvalue.SetSizeExact(2),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: target.create (new target creation)
// ---------------------------------------------------------------------------

// TestAccSyncResourceTargetCreate verifies creating a new target table via
// target.create instead of referencing an existing target.object.
func TestAccSyncResourceTargetCreate(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncCreate-%s", uuid.NewString())
	apiKey := APIKey()
	tableName := fmt.Sprintf("test_create_%s", strings.ReplaceAll(uuid.NewString()[:8], "-", "_"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:   name,
					APIKey: apiKey,
					Mode:   "replace",
					Active: "false",
					Fields: defaultAdvancedSyncFields,
					Target: fmt.Sprintf(`{
    connection_id = polytomic_postgresql_connection.test.id
    create = {
      "name" = "public.%s"
    }
  }`, tableName),
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("mode"),
						knownvalue.StringExact("replace"),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: target_filters (target-side filters, requires mode=update + identity)
// ---------------------------------------------------------------------------

func TestAccSyncResourceTargetFilters(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncTgtFilter-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:       name,
					APIKey:     apiKey,
					Mode:       "update",
					Active:     "false",
					ModelQuery: identityModelQuery,
					Fields:     identitySyncFields,
					Identity: `{
      source = {
        model_id = polytomic_model.test.id
        field    = "email"
      }
      target   = "email"
      function = "Equality"
    }`,
					TargetFilters: `[
    {
      field    = "email"
      function = "IsNotNull"
    }
  ]`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("target_filters"),
						knownvalue.SetSizeExact(1),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Daily schedule
// ---------------------------------------------------------------------------

func TestAccSyncResourceScheduleDaily(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncSchedDaily-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:   name,
					APIKey: apiKey,
					Mode:   "replace",
					Active: "false",
					Fields: defaultAdvancedSyncFields,
					Schedule: `{
    frequency = "daily"
    hour      = "10"
    minute    = "30"
  }`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("schedule").AtMapKey("frequency"),
						knownvalue.StringExact("daily"),
					),
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("schedule").AtMapKey("hour"),
						knownvalue.StringExact("10"),
					),
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("schedule").AtMapKey("minute"),
						knownvalue.StringExact("30"),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Update lifecycle (create, then modify mode/identity/flags)
// ---------------------------------------------------------------------------

func TestAccSyncResourceUpdateLifecycle(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncUpdate-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with replace mode using multi-column model.
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:       name,
					APIKey:     apiKey,
					Mode:       "replace",
					Active:     "false",
					ModelQuery: identityModelQuery,
					Fields:     identitySyncFields,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
					testAccSyncMode(t, apiKey, "replace"),
				),
			},
			// Step 2: Change to updateOrCreate mode with identity.
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:       name,
					APIKey:     apiKey,
					Mode:       "updateOrCreate",
					Active:     "false",
					ModelQuery: identityModelQuery,
					Fields:     identitySyncFields,
					Identity: `{
      source = {
        model_id = polytomic_model.test.id
        field    = "email"
      }
      target   = "email"
      function = "Equality"
    }`,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccSyncMode(t, apiKey, "updateOrCreate"),
					testAccSyncHasIdentity(t, apiKey),
				),
			},
			// Step 3: Add sync_all_records flag.
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:       name,
					APIKey:     apiKey,
					Mode:       "updateOrCreate",
					Active:     "false",
					ModelQuery: identityModelQuery,
					Fields:     identitySyncFields,
					Identity: `{
      source = {
        model_id = polytomic_model.test.id
        field    = "email"
      }
      target   = "email"
      function = "Equality"
    }`,
					SyncAllRecords: "true",
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("sync_all_records"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Append mode (no identity needed)
// ---------------------------------------------------------------------------

// TestAccSyncResourceModeAppend is skipped because the PostgreSQL target
// does not support append mode. This would need a different target type (e.g., S3/CSV).
func TestAccSyncResourceModeAppend(t *testing.T) {
	t.Skip("Skipped: PostgreSQL target does not support append mode")
}

// ---------------------------------------------------------------------------
// Test: Create mode (with identity for dedup)
// ---------------------------------------------------------------------------

func TestAccSyncResourceModeCreate(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncCreate-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:       name,
					APIKey:     apiKey,
					Mode:       "create",
					Active:     "false",
					ModelQuery: identityModelQuery,
					Fields:     identitySyncFields,
					Identity: `{
      source = {
        model_id = polytomic_model.test.id
        field    = "email"
      }
      target   = "email"
      function = "Equality"
    }`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("mode"),
						knownvalue.StringExact("create"),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
					testAccSyncMode(t, apiKey, "create"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Encryption passphrase with field-level encryption
// ---------------------------------------------------------------------------

func TestAccSyncResourceEncryption(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncEncrypt-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:                 name,
					APIKey:               apiKey,
					Mode:                 "replace",
					Active:               "false",
					EncryptionPassphrase: "test-passphrase-12345",
					Fields: `[
    {
      source = {
        field    = "email"
        model_id = polytomic_model.test.id
      }
      target             = "email"
      encryption_enabled = true
    }
  ]`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Field with sync_mode override
// ---------------------------------------------------------------------------

func TestAccSyncResourceFieldSyncMode(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncFieldMode-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:       name,
					APIKey:     apiKey,
					Mode:       "updateOrCreate",
					Active:     "false",
					ModelQuery: identityModelQuery,
					Fields: `[
    {
      source = {
        field    = "name"
        model_id = polytomic_model.test.id
      }
      target    = "name"
      sync_mode = "create"
    }
  ]`,
					Identity: `{
      source = {
        model_id = polytomic_model.test.id
        field    = "email"
      }
      target   = "email"
      function = "Equality"
    }`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("fields"),
						knownvalue.SetSizeExact(1),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: Schedule frequencies — hourly, weekly, continuous
// ---------------------------------------------------------------------------

func TestAccSyncResourceScheduleHourly(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncSchedHr-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:   name,
					APIKey: apiKey,
					Mode:   "replace",
					Active: "false",
					Fields: defaultAdvancedSyncFields,
					Schedule: `{
    frequency = "hourly"
    minute    = "15"
  }`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("schedule").AtMapKey("frequency"),
						knownvalue.StringExact("hourly"),
					),
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("schedule").AtMapKey("minute"),
						knownvalue.StringExact("15"),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
				),
			},
		},
	})
}

func TestAccSyncResourceScheduleWeekly(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncSchedWk-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:   name,
					APIKey: apiKey,
					Mode:   "replace",
					Active: "false",
					Fields: defaultAdvancedSyncFields,
					Schedule: `{
    frequency   = "weekly"
    day_of_week = "1"
    hour        = "8"
    minute      = "0"
  }`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("schedule").AtMapKey("frequency"),
						knownvalue.StringExact("weekly"),
					),
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("schedule").AtMapKey("day_of_week"),
						knownvalue.StringExact("1"),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
				),
			},
		},
	})
}

func TestAccSyncResourceScheduleContinuous(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncSchedCont-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncAdvancedTestConfig(t, syncAdvancedTestArgs{
					Name:   name,
					APIKey: apiKey,
					Mode:   "replace",
					Active: "false",
					Fields: defaultAdvancedSyncFields,
					Schedule: `{
    frequency = "continuous"
  }`,
				}),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.test",
						tfjsonpath.New("schedule").AtMapKey("frequency"),
						knownvalue.StringExact("continuous"),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccSyncExists(t, name, apiKey),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Test: schedule.run_after (run this sync after another completes)
// ---------------------------------------------------------------------------

func TestAccSyncResourceScheduleRunAfter(t *testing.T) {
	name := fmt.Sprintf("TestAccSyncRunAfter-%s", uuid.NewString())
	apiKey := APIKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: syncRunAfterTestConfig(t, name, apiKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_sync.dependent",
						tfjsonpath.New("schedule").AtMapKey("frequency"),
						knownvalue.StringExact("runafter"),
					),
				},
			},
		},
	})
}

func syncRunAfterTestConfig(t *testing.T, name string, apiKey bool) string {
	t.Helper()
	tmpl := texttemplate.Must(texttemplate.New("run-after").Parse(syncRunAfterTemplate))
	var buf strings.Builder
	require.NoError(t, tmpl.Execute(&buf, struct {
		Name     string
		APIKey   bool
		Postgres postgresTestConfig
	}{
		Name:     name,
		APIKey:   apiKey,
		Postgres: testPostgresConfig(t),
	}))
	return buf.String()
}

const syncRunAfterTemplate = `
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

resource "polytomic_sync" "primary" {
  name   = "{{.Name}}-primary"
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

resource "polytomic_sync" "dependent" {
  name   = "{{.Name}}-dependent"
  mode   = "replace"
  active = false
  schedule = {
    frequency = "runafter"
    run_after = {
      sync_ids = [polytomic_sync.primary.id]
    }
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
