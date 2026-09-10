package importer

import (
	"bytes"
	"context"
	"reflect"
	"regexp"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/polytomic/polytomic-go/v25"
)

func TestSchemaOverridesAdd(t *testing.T) {
	yes, no := pointer.To(true), pointer.To(false)
	stringType := polytomic.UtilFieldTypeString

	t.Run("records overridden keys and user-managed fields", func(t *testing.T) {
		s := NewSchemaOverrides(nil, "org")
		ok := s.add("Warehouse", "conn-1", []*polytomic.Schema{
			{ID: pointer.To("public.orders"), Fields: []*polytomic.SchemaField{
				{ID: pointer.To("id"), IsPrimaryKey: no, SourcePrimaryKey: yes, PrimaryKeyOverride: no},
				{ID: pointer.To("order_no"), IsPrimaryKey: yes, SourcePrimaryKey: no, PrimaryKeyOverride: yes},
				{ID: pointer.To("tenant_id"), IsPrimaryKey: yes, SourcePrimaryKey: yes},
			}},
			// Detected keys only, so no primary keys resource.
			{ID: pointer.To("public.users"), Fields: []*polytomic.SchemaField{
				{ID: pointer.To("id"), IsPrimaryKey: yes, SourcePrimaryKey: yes},
			}},
			{ID: pointer.To("public.events"), Fields: []*polytomic.SchemaField{
				{ID: pointer.To("city"), Name: pointer.To("City"), Type: &stringType, UserManaged: yes, SourcePrimaryKey: no},
			}},
			// Overrides that unmark every key cannot be expressed by the resource.
			{ID: pointer.To("public.logs"), Fields: []*polytomic.SchemaField{
				{ID: pointer.To("id"), IsPrimaryKey: no, SourcePrimaryKey: yes, PrimaryKeyOverride: no},
			}},
		})
		if !ok {
			t.Fatal("add reported that overrides are unsupported")
		}

		wantKeys := map[string]schemaPrimaryKeys{
			"warehouse_public_orders": {ConnectionID: "conn-1", SchemaID: "public.orders", FieldIDs: []string{"order_no", "tenant_id"}},
		}
		if !reflect.DeepEqual(s.PrimaryKeys, wantKeys) {
			t.Errorf("primary keys: got %+v, want %+v", s.PrimaryKeys, wantKeys)
		}
		if got := sortedKeys(s.Fields); !reflect.DeepEqual(got, []string{"warehouse_public_events_city"}) {
			t.Errorf("fields: got %v", got)
		}
	})

	t.Run("skips primary keys the deployment does not report", func(t *testing.T) {
		s := NewSchemaOverrides(nil, "org")
		ok := s.add("Legacy", "conn-2", []*polytomic.Schema{
			{ID: pointer.To("orders"), Fields: []*polytomic.SchemaField{
				{ID: pointer.To("id"), IsPrimaryKey: yes},
				{ID: pointer.To("city"), Name: pointer.To("City"), Type: &stringType, UserManaged: yes},
			}},
		})
		if ok {
			t.Error("want add to report that overrides are unsupported")
		}
		if len(s.PrimaryKeys) != 0 {
			t.Errorf("primary keys: got %+v", s.PrimaryKeys)
		}
		if len(s.Fields) != 1 {
			t.Errorf("fields: got %+v", s.Fields)
		}
	})

	t.Run("suffixes colliding names", func(t *testing.T) {
		s := NewSchemaOverrides(nil, "org")
		schemas := []*polytomic.Schema{{ID: pointer.To("orders"), Fields: []*polytomic.SchemaField{
			{ID: pointer.To("city"), Name: pointer.To("City"), Type: &stringType, UserManaged: yes, SourcePrimaryKey: no},
		}}}
		s.add("Shop", "conn-1", schemas)
		s.add("Shop", "conn-2", schemas)
		if got := sortedKeys(s.Fields); !reflect.DeepEqual(got, []string{"shop_orders_city", "shop_orders_city_2"}) {
			t.Errorf("fields: got %v", got)
		}
	})
}

func TestSchemaOverridesGenerate(t *testing.T) {
	yes, no := pointer.To(true), pointer.To(false)
	stringType := polytomic.UtilFieldTypeString
	numberType := polytomic.UtilFieldTypeNumber
	arrayType := polytomic.UtilFieldTypeArray
	const connectionID = "0b6f3a52-7d1e-4c8a-9f2b-3e5d7c9a1b24"

	s := NewSchemaOverrides(nil, "org-1")
	s.add("Orders", connectionID, []*polytomic.Schema{
		{ID: pointer.To("shop.orders"), Fields: []*polytomic.SchemaField{
			{ID: pointer.To("id"), IsPrimaryKey: no, SourcePrimaryKey: yes, PrimaryKeyOverride: no},
			{ID: pointer.To("order_no"), IsPrimaryKey: yes, SourcePrimaryKey: no, PrimaryKeyOverride: yes},
			{ID: pointer.To("city"), Name: pointer.To("City"), Type: &stringType, Path: pointer.To("$.address.city"), UserManaged: yes, SourcePrimaryKey: no},
			{ID: pointer.To("total"), Name: pointer.To("Total"), Type: &numberType, UserManaged: yes, SourcePrimaryKey: no,
				TypeSpec: pointer.To[any]([]any{"decimal", map[string]any{"precision": 12.0, "scale": 2.0}})},
			{ID: pointer.To("tags"), Name: pointer.To("Tags"), Type: &arrayType, UserManaged: yes, SourcePrimaryKey: no,
				TypeSpec: pointer.To[any]([]any{"array", "string"})},
		}},
		{ID: pointer.To("exports/2026 orders.csv"), Fields: []*polytomic.SchemaField{
			{ID: pointer.To("total"), Name: pointer.To("total"), Type: &stringType, UserManaged: yes, SourcePrimaryKey: no},
		}},
	})

	var tf, imports bytes.Buffer
	refs := map[string]string{connectionID: "polytomic_mongodb_connection.orders.id"}
	if err := s.GenerateTerraformFiles(context.Background(), &tf, refs); err != nil {
		t.Fatal(err)
	}
	if err := s.GenerateImports(context.Background(), &imports); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		`resource "polytomic_connection_schema_primary_keys" "orders_shop_orders" {`,
		`field_ids\s+= \["order_no"\]`,
		`resource "polytomic_connection_schema_field" "orders_shop_orders_city" {`,
		`connection_id\s+= polytomic_mongodb_connection\.orders\.id`,
		`organization\s+= local\.organization_id`,
		`path\s+= "\$\.address\.city"`,
		`type\s+= "decimal"`,
		`precision\s+= 12`,
		`scale\s+= 2`,
		`type_spec\s+= jsonencode\(\["array", "string"\]\)`,
	} {
		if !regexp.MustCompile(want).Match(tf.Bytes()) {
			t.Errorf("missing %s in:\n%s", want, tf.String())
		}
	}
	if bytes.Contains(tf.Bytes(), []byte(connectionID)) {
		t.Errorf("connection ID was not replaced with a reference:\n%s", tf.String())
	}

	wantImports := "terraform import polytomic_connection_schema_primary_keys.orders_shop_orders org-1/" + connectionID + "/shop.orders\n" +
		"terraform import polytomic_connection_schema_field.orders_exports_2026_orders_csv_total 'org-1/" + connectionID + "/exports/2026 orders.csv/total'\n" +
		"terraform import polytomic_connection_schema_field.orders_shop_orders_city org-1/" + connectionID + "/shop.orders/city\n" +
		"terraform import polytomic_connection_schema_field.orders_shop_orders_tags org-1/" + connectionID + "/shop.orders/tags\n" +
		"terraform import polytomic_connection_schema_field.orders_shop_orders_total org-1/" + connectionID + "/shop.orders/total\n"
	if got := imports.String(); got != wantImports {
		t.Errorf("imports: got\n%s\nwant\n%s", got, wantImports)
	}
}
