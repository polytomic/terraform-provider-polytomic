package connections

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//go:embed force_destroy.md
var forceDestroyMessage string

type connectionData struct {
	Organization  types.String `tfsdk:"organization"`
	Name          types.String `tfsdk:"name"`
	Id            types.String `tfsdk:"id"`
	Configuration types.Object `tfsdk:"configuration"`
	ForceDestroy  types.Bool   `tfsdk:"force_destroy"`
}

func getValueOrEmpty(v any, typ string) any {
	switch typ {
	case "string":
		if v == nil {
			return ""
		}
		return v.(string)
	case "bool":
		if v == nil {
			return false
		}
		return v.(bool)
	case "int":
		if v == nil {
			return 0
		}
		return v.(int)
	case "float64":
		if v == nil {
			return 0.0
		}
		return v.(float64)
	case "int64":
		if v == nil {
			return int64(0)
		}
		return v.(int64)
	default:
		panic(fmt.Sprintf("unsupported type %s", typ))
	}
}

func objectMapValue(ctx context.Context, value types.Object) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	for k, v := range value.Attributes() {
		val, err := attrValue(ctx, v)
		if err != nil {
			return nil, fmt.Errorf("error converting value for %s: %w", k, err)
		}
		out[k] = val
	}

	return out, nil
}

func attrValue(ctx context.Context, val attr.Value) (interface{}, error) {
	switch tv := val.(type) {
	case types.Bool:
		return tv.ValueBool(), nil
	case types.Int32:
		return tv.ValueInt32(), nil
	case types.Int64:
		return tv.ValueInt64(), nil
	case types.Float32:
		return tv.ValueFloat32(), nil
	case types.Float64:
		return tv.ValueFloat64(), nil
	case types.Number:
		return tv.ValueBigFloat(), nil
	case types.String:
		return tv.ValueString(), nil
	case types.Object:
		return objectMapValue(ctx, tv)
	case types.Set:
		elemsIn := tv.Elements()
		elemsOut := make([]interface{}, len(elemsIn))
		for i, elem := range elemsIn {
			elemOut, err := attrValue(ctx, elem)
			if err != nil {
				return nil, fmt.Errorf("error converting set element %d: %w", i, err)
			}
			elemsOut[i] = elemOut
		}
		return elemsOut, nil
	}

	return nil, fmt.Errorf("unsupported type %T", val)
}
