package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable                    = typeSpecType{}
	_ basetypes.StringValuableWithSemanticEquals = typeSpecValue{}
)

// typeSpecType is the type of polytomic_connection_schema_field's type_spec:
// JSON whose values are semantically equal when they describe the same field
// type. The API fills in defaults when it stores a definition, such as the
// unit of a sized string, so a configured definition that leaves them out
// still matches the one the API returns.
type typeSpecType struct {
	basetypes.StringType
}

func (t typeSpecType) String() string {
	return "provider.typeSpecType"
}

func (t typeSpecType) ValueType(ctx context.Context) attr.Value {
	return typeSpecValue{}
}

func (t typeSpecType) Equal(o attr.Type) bool {
	_, ok := o.(typeSpecType)
	return ok
}

func (t typeSpecType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return typeSpecValue{Normalized: jsontypes.Normalized{StringValue: in}}, nil
}

func (t typeSpecType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	v, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	s, ok := v.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", v)
	}
	return typeSpecValue{Normalized: jsontypes.Normalized{StringValue: s}}, nil
}

// typeSpecValue is a type_spec value. It keeps jsontypes.Normalized's JSON
// validation.
type typeSpecValue struct {
	jsontypes.Normalized
}

func newTypeSpecNull() typeSpecValue {
	return typeSpecValue{Normalized: jsontypes.NewNormalizedNull()}
}

func newTypeSpecUnknown() typeSpecValue {
	return typeSpecValue{Normalized: jsontypes.NewNormalizedUnknown()}
}

func newTypeSpecValue(s string) typeSpecValue {
	return typeSpecValue{Normalized: jsontypes.NewNormalizedValue(s)}
}

func (v typeSpecValue) Type(context.Context) attr.Type {
	return typeSpecType{}
}

func (v typeSpecValue) Equal(o attr.Value) bool {
	other, ok := o.(typeSpecValue)
	return ok && v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals reports whether both values describe the same field
// type once the defaults the API fills in are applied.
func (v typeSpecValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	newValue, ok := newValuable.(typeSpecValue)
	if !ok {
		diags.AddError("Semantic Equality Check Error", fmt.Sprintf(
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\nExpected Value Type: %T\nGot Value Type: %T", v, newValuable))
		return false, diags
	}

	var specs [2]any
	for i, s := range []string{v.ValueString(), newValue.ValueString()} {
		if err := json.Unmarshal([]byte(s), &specs[i]); err != nil {
			diags.AddError("Semantic Equality Check Error", "Invalid type_spec JSON: "+err.Error())
			return false, diags
		}
	}
	return reflect.DeepEqual(canonicalTypeSpec(specs[0]), canonicalTypeSpec(specs[1])), diags
}

// canonicalTypeSpec returns a decoded definition as the API stores it: a sized
// string's unit defaults to characters, and a string without a length is a
// plain string. Other definitions are returned unchanged, apart from the
// definitions nested in them.
func canonicalTypeSpec(spec any) any {
	v, ok := spec.([]any)
	if !ok || len(v) != 2 {
		return spec
	}
	switch v[0] {
	case "array", "map":
		return []any{v[0], canonicalTypeSpec(v[1])}
	case "object":
		if attrs, ok := v[1].(map[string]any); ok {
			canonical := make(map[string]any, len(attrs))
			for k, a := range attrs {
				canonical[k] = canonicalTypeSpec(a)
			}
			return []any{v[0], canonical}
		}
	case "versioned":
		if versions, ok := v[1].([]any); ok {
			canonical := make([]any, len(versions))
			for i, t := range versions {
				canonical[i] = canonicalTypeSpec(t)
			}
			return []any{v[0], canonical}
		}
	case "string":
		if details, ok := v[1].(map[string]any); ok {
			length, _ := jsonInt(details["length"])
			if length <= 0 {
				return "string"
			}
			unit, _ := details["unit"].(string)
			if unit == "" {
				unit = "characters"
			}
			return []any{v[0], map[string]any{"length": length, "unit": unit}}
		}
	}
	return spec
}
