package provider

import (
	"encoding/json"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/polytomic/polytomic-go/v25"
)

// decodeJSON round-trips v through JSON, the form definitions arrive in.
func decodeJSON(t *testing.T, v any) any {
	t.Helper()
	buf, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err := json.Unmarshal(buf, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestFieldTypeRoundTrip(t *testing.T) {
	if len(fieldTypeNames) != len(basicFieldTypes)+len(detailedFieldTypes) {
		t.Fatalf("fieldTypeNames has %d names, the type maps have %d", len(fieldTypeNames), len(basicFieldTypes)+len(detailedFieldTypes))
	}

	for _, name := range fieldTypeNames {
		t.Run(name, func(t *testing.T) {
			basic, spec, err := fieldTypeSpec(name, 12, 2)
			if err != nil {
				t.Fatal(err)
			}
			decoded := decodeJSON(t, spec)

			got, precision, scale, ok := fieldTypeFromSpec(decoded)
			if !ok || got != name {
				t.Errorf("fieldTypeFromSpec(%v) = %q, %v", decoded, got, ok)
			}
			if name == "decimal" {
				if pointer.Get(precision) != 12 || pointer.Get(scale) != 2 {
					t.Errorf("decimal: got precision %v scale %v", precision, scale)
				}
			} else if precision != nil || scale != nil {
				t.Errorf("unexpected precision %v scale %v", precision, scale)
			}

			if gotBasic, err := specBasicType(decoded); err != nil || gotBasic != basic {
				t.Errorf("specBasicType(%v) = %q, %v; want %q", decoded, gotBasic, err, basic)
			}

			def, err := newTypesDefinition(spec)
			if err != nil {
				t.Fatal(err)
			}
			sent, err := json.Marshal(def)
			if err != nil {
				t.Fatal(err)
			}
			want, _ := json.Marshal(spec)
			if string(sent) != string(want) {
				t.Errorf("request definition: got %s, want %s", sent, want)
			}
		})
	}
}

func TestFieldTypeFromSpecWithoutName(t *testing.T) {
	for _, tc := range []struct {
		spec  string
		basic string
	}{
		{spec: `["array", "string"]`, basic: "array"},
		{spec: `["string", {"length": 12}]`, basic: "string"},
		{spec: `["map", "number"]`, basic: "object"},
		{spec: `["object", {"city": "string"}]`, basic: "object"},
		{spec: `["versioned", [["decimal", {"precision": 10, "scale": 2}], "number"]]`, basic: "number"},
	} {
		t.Run(tc.spec, func(t *testing.T) {
			var spec any
			if err := json.Unmarshal([]byte(tc.spec), &spec); err != nil {
				t.Fatal(err)
			}
			if name, _, _, ok := fieldTypeFromSpec(spec); ok {
				t.Errorf("want no type name, got %q", name)
			}
			if basic, err := specBasicType(spec); err != nil || basic != tc.basic {
				t.Errorf("specBasicType = %q, %v; want %q", basic, err, tc.basic)
			}
		})
	}

	for _, bad := range []any{nil, []any{}, 5.0, "varchar"} {
		if _, err := newTypesDefinition(bad); err == nil {
			t.Errorf("newTypesDefinition(%v): want error", bad)
		}
	}
}

func TestSchemaFieldTypeAttributes(t *testing.T) {
	number := polytomic.UtilFieldTypeNumber
	array := polytomic.UtilFieldTypeArray

	name, _, _, spec, err := SchemaFieldTypeAttributes(&polytomic.SchemaField{Type: &number, TypeSpec: pointer.To[any]("bigint")})
	if err != nil || name != "bigint" || spec != "" {
		t.Errorf("bigint: got %q %q %v", name, spec, err)
	}

	name, _, _, spec, err = SchemaFieldTypeAttributes(&polytomic.SchemaField{Type: &array, TypeSpec: pointer.To[any]([]any{"array", "string"})})
	if err != nil || name != "" || spec != `["array","string"]` {
		t.Errorf("array of string: got %q %q %v", name, spec, err)
	}

	name, _, _, spec, err = SchemaFieldTypeAttributes(&polytomic.SchemaField{Type: &number})
	if err != nil || name != "number" || spec != "" {
		t.Errorf("no type_spec: got %q %q %v", name, spec, err)
	}
}
