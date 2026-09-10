package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestParseSchemaFieldID(t *testing.T) {
	for _, tc := range []struct {
		id                             string
		org, connection, schema, field string
		wantErr                        bool
	}{
		{id: "org/conn/shop.orders/city", org: "org", connection: "conn", schema: "shop.orders", field: "city"},
		{id: "org/conn/exports/2026/orders.csv/city", org: "org", connection: "conn", schema: "exports/2026/orders.csv", field: "city"},
		{id: "org/conn/orders/price%2Funit", org: "org", connection: "conn", schema: "orders", field: "price/unit"},
		{id: "org/conn/orders/price%2funit", org: "org", connection: "conn", schema: "orders", field: "price/unit"},
		{id: "org/conn/exports/orders.csv/100%25", org: "org", connection: "conn", schema: "exports/orders.csv", field: "100%"},
		// An unescaped "%" is kept.
		{id: "org/conn/orders/pct%", org: "org", connection: "conn", schema: "orders", field: "pct%"},
		{id: "org/conn/shop.orders", wantErr: true},
		{id: "org/conn//city", wantErr: true},
		{id: "/conn/shop.orders/city", wantErr: true},
	} {
		t.Run(tc.id, func(t *testing.T) {
			org, connection, schema, field, err := parseSchemaFieldID(tc.id)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got %q %q %q %q", org, connection, schema, field)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if org != tc.org || connection != tc.connection || schema != tc.schema || field != tc.field {
				t.Errorf("got %q %q %q %q", org, connection, schema, field)
			}
		})
	}
}

func TestSchemaFieldResourceIDRoundTrip(t *testing.T) {
	const schemaID = "exports/2026/orders.csv"
	for _, fieldID := range []string{"city", "price/unit", "100%", "a%2Fb", "%252F", "/"} {
		t.Run(fieldID, func(t *testing.T) {
			id := SchemaFieldResourceID("org", "conn", schemaID, fieldID)
			org, connection, schema, field, err := parseSchemaFieldID(id)
			if err != nil {
				t.Fatal(err)
			}
			if org != "org" || connection != "conn" || schema != schemaID || field != fieldID {
				t.Errorf("%s: got %q %q %q %q", id, org, connection, schema, field)
			}
		})
	}
}

func TestPlanFieldType(t *testing.T) {
	ctx := context.Background()
	state := connectionSchemaFieldResourceModel{
		Type:      types.StringValue("number"),
		Precision: types.Int64Null(),
		Scale:     types.Int64Null(),
		TypeSpec:  newTypeSpecValue(`"number"`),
	}
	// unset returns a configuration that sets none of the type attributes.
	unset := func() connectionSchemaFieldResourceModel {
		return connectionSchemaFieldResourceModel{
			Type:      types.StringNull(),
			Precision: types.Int64Null(),
			Scale:     types.Int64Null(),
			TypeSpec:  newTypeSpecNull(),
		}
	}
	// planFor mimics the framework: configured values carry over, and
	// unconfigured computed values are unknown.
	planFor := func(config connectionSchemaFieldResourceModel) connectionSchemaFieldResourceModel {
		plan := config
		if config.Type.IsNull() {
			plan.Type = types.StringUnknown()
		}
		if config.Precision.IsNull() {
			plan.Precision = types.Int64Unknown()
		}
		if config.Scale.IsNull() {
			plan.Scale = types.Int64Unknown()
		}
		if config.TypeSpec.IsNull() {
			plan.TypeSpec = newTypeSpecUnknown()
		}
		return plan
	}

	t.Run("keeps state while the type is unchanged", func(t *testing.T) {
		config := unset()
		plan := planFor(config)
		planFieldType(ctx, config, state, &plan)
		if !plan.Type.Equal(state.Type) || !plan.TypeSpec.Equal(state.TypeSpec) || !plan.Precision.IsNull() || !plan.Scale.IsNull() {
			t.Errorf("got %+v", plan)
		}
	})

	t.Run("a new type name makes type_spec unknown", func(t *testing.T) {
		config := unset()
		config.Type = types.StringValue("bigint")
		plan := planFor(config)
		planFieldType(ctx, config, state, &plan)
		if !plan.TypeSpec.IsUnknown() || !plan.Precision.IsNull() || !plan.Scale.IsNull() {
			t.Errorf("got %+v", plan)
		}
	})

	t.Run("a new type_spec makes the other type attributes unknown", func(t *testing.T) {
		config := unset()
		config.TypeSpec = newTypeSpecValue(`["decimal", {"precision": 12, "scale": 2}]`)
		plan := planFor(config)
		planFieldType(ctx, config, state, &plan)
		if !plan.Type.IsUnknown() || !plan.Precision.IsUnknown() || !plan.Scale.IsUnknown() {
			t.Errorf("got %+v", plan)
		}
	})

	t.Run("an equivalent type_spec is not a change", func(t *testing.T) {
		config := unset()
		config.TypeSpec = newTypeSpecValue(` "number" `)
		plan := planFor(config)
		planFieldType(ctx, config, state, &plan)
		if !plan.Type.Equal(state.Type) {
			t.Errorf("got %+v", plan)
		}
	})

	t.Run("a type_spec without the defaults the API fills in is not a change", func(t *testing.T) {
		state := state
		state.Type = types.StringNull()
		state.TypeSpec = newTypeSpecValue(`["string",{"length":12,"unit":"characters"}]`)
		config := unset()
		config.TypeSpec = newTypeSpecValue(`["string", {"length": 12}]`)
		plan := planFor(config)
		planFieldType(ctx, config, state, &plan)
		if !plan.Type.Equal(state.Type) || !plan.Precision.IsNull() {
			t.Errorf("got %+v", plan)
		}
	})
}

func TestTypeSpecSemanticEquals(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{`["string", {"length": 12}]`, `["string",{"length":12,"unit":"characters"}]`, true},
		{`["string", {"length": 12, "unit": "bytes"}]`, `["string",{"length":12,"unit":"characters"}]`, false},
		{`["string", {"length": 12}]`, `["string",{"length":13,"unit":"characters"}]`, false},
		{`["string", {}]`, `"string"`, true},
		{`["array", ["string", {"length": 3}]]`, `["array",["string",{"length":3,"unit":"characters"}]]`, true},
		{`["map", ["string", {"length": 3}]]`, `["map",["string",{"length":3,"unit":"characters"}]]`, true},
		{`["object", {"b": "int", "a": ["string", {"length": 3}]}]`, `["object",{"a": ["string",{"length":3,"unit":"characters"}], "b": "int"}]`, true},
		{`["object", {"a": ["string", {"length": 3}]}]`, `["object",{"a": "string"}]`, false},
		{`["versioned", [["string", {"length": 3}], "string"]]`, `["versioned",[["string",{"length":3,"unit":"characters"}],"string"]]`, true},
		{` "number" `, `"number"`, true},
		{`"int"`, `"bigint"`, false},
	} {
		t.Run(tc.a+" "+tc.b, func(t *testing.T) {
			got, diags := newTypeSpecValue(tc.a).StringSemanticEquals(context.Background(), newTypeSpecValue(tc.b))
			if diags.HasError() {
				t.Fatal(diags)
			}
			if got != tc.want {
				t.Errorf("got %t, want %t", got, tc.want)
			}
		})
	}
}

func TestAccConnectionSchemaField(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("%s must be set for acceptance tests", resource.EnvTfAcc)
	}
	name := fmt.Sprintf("TestAccSchemaField-%s", uuid.NewString())
	mongo := testMongoConfig(t)
	args := func(cityLabel, amountType string) TestCaseTfArgs {
		return TestCaseTfArgs{
			Name:   name,
			APIKey: APIKey(),
			Extra:  map[string]any{"Mongo": mongo, "CityLabel": cityLabel, "AmountType": amountType},
		}
	}
	orders := "data.polytomic_connection_schema.orders"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, connectionSchemaFieldTemplate, args("City", "number")),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_connection_schema_field.city",
						tfjsonpath.New("path"), knownvalue.StringExact("$.address.city")),
					statecheck.ExpectKnownValue("polytomic_connection_schema_field.amount",
						tfjsonpath.New("label"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(orders,
						tfjsonpath.New("fields_by_id").AtMapKey("amount"),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"type":         knownvalue.StringExact("number"),
							"user_managed": knownvalue.Bool(true),
						})),
					statecheck.ExpectKnownValue(orders,
						tfjsonpath.New("fields_by_id").AtMapKey("city"),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"name":         knownvalue.StringExact("City"),
							"path":         knownvalue.StringExact("$.address.city"),
							"user_managed": knownvalue.Bool(true),
						})),
				},
			},
			{
				Config: TestCaseTfResource(t, connectionSchemaFieldTemplate, args("Town", "decimal")),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_connection_schema_field.city",
						tfjsonpath.New("label"), knownvalue.StringExact("Town")),
					statecheck.ExpectKnownValue("polytomic_connection_schema_field.amount",
						tfjsonpath.New("type_spec"), knownvalue.StringExact(`["decimal",{"precision":12,"scale":2}]`)),
					statecheck.ExpectKnownValue("polytomic_connection_schema_field.tags",
						tfjsonpath.New("type"), knownvalue.StringExact("array")),
					// The API adds the default unit; state keeps the configured value.
					statecheck.ExpectKnownValue("polytomic_connection_schema_field.code",
						tfjsonpath.New("type_spec"), knownvalue.StringExact(`["string",{"length":12}]`)),
				},
			},
			{
				ResourceName:      "polytomic_connection_schema_field.city",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["polytomic_connection_schema_field.city"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return rs.Primary.Attributes["id"], nil
				},
			},
		},
	})
}

const connectionSchemaFieldTemplate = `
{{if not .APIKey}}
resource "polytomic_organization" "test" {
  name = "{{.Name}}"
}
{{end}}

resource "polytomic_mongodb_connection" "test" {
  name = "{{.Name}}"
  configuration = {
    hosts    = "{{.Extra.Mongo.Hosts}}"
    database = "{{.Extra.Mongo.Database}}"
    username = "{{.Extra.Mongo.Username}}"
    password = "{{.Extra.Mongo.Password}}"
    params   = "authSource=admin"
  }
{{if not .APIKey}}
  organization = polytomic_organization.test.id
{{end}}
}

data "polytomic_connection_schemas" "test" {
  connection_id = polytomic_mongodb_connection.test.id
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}

locals {
  orders_schema = one([for s in data.polytomic_connection_schemas.test.schemas : s.id if endswith(s.id, "orders")])
}

resource "polytomic_connection_schema_field" "city" {
  connection_id = polytomic_mongodb_connection.test.id
  schema_id     = local.orders_schema
  field_id      = "city"
  label         = "{{.Extra.CityLabel}}"
  type          = "string"
  path          = "$.address.city"
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}

resource "polytomic_connection_schema_field" "amount" {
  connection_id = polytomic_mongodb_connection.test.id
  schema_id     = local.orders_schema
  field_id      = "amount"
  type          = "{{.Extra.AmountType}}"
{{if eq .Extra.AmountType "decimal"}}
  precision     = 12
  scale         = 2
{{end}}
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}

resource "polytomic_connection_schema_field" "tags" {
  connection_id = polytomic_mongodb_connection.test.id
  schema_id     = local.orders_schema
  field_id      = "tag_list"
  label         = "Tags"
  path          = "$.tags"
  type_spec     = jsonencode(["array", "string"])
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}

resource "polytomic_connection_schema_field" "code" {
  connection_id = polytomic_mongodb_connection.test.id
  schema_id     = local.orders_schema
  field_id      = "code"
  label         = "Code"
  type_spec     = jsonencode(["string", { length = 12 }])
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
}

data "polytomic_connection_schema" "orders" {
  connection_id = polytomic_mongodb_connection.test.id
  schema_id     = local.orders_schema
{{if not .APIKey}}
  organization  = polytomic_organization.test.id
{{end}}
  depends_on = [
    polytomic_connection_schema_field.city,
    polytomic_connection_schema_field.amount,
    polytomic_connection_schema_field.tags,
    polytomic_connection_schema_field.code,
  ]
}
`
