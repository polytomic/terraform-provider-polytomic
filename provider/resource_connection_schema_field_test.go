package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
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

func TestAccConnectionSchemaField(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("%s must be set for acceptance tests", resource.EnvTfAcc)
	}
	name := fmt.Sprintf("TestAccSchemaField-%s", uuid.NewString())
	mongo := testMongoConfig(t)
	args := func(cityLabel string) TestCaseTfArgs {
		return TestCaseTfArgs{
			Name:   name,
			APIKey: APIKey(),
			Extra:  map[string]any{"Mongo": mongo, "CityLabel": cityLabel},
		}
	}
	orders := "data.polytomic_connection_schema.orders"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestCaseTfResource(t, connectionSchemaFieldTemplate, args("City")),
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
				Config: TestCaseTfResource(t, connectionSchemaFieldTemplate, args("Town")),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("polytomic_connection_schema_field.city",
						tfjsonpath.New("label"), knownvalue.StringExact("Town")),
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
  type          = "number"
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
  ]
}
`
