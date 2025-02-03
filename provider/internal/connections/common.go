package connections

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	// "go.opentelemetry.io/otel/attribute"
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
		if v.IsNull() || v.IsUnknown() {
			// don't want to write empty values
			continue
		}

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

func clearSensitiveValuesFromRead(attrs map[string]schema.Attribute, config map[string]any) map[string]any {
	for k, v := range attrs {
		switch v := v.(type) {
		case schema.ListNestedAttribute:
			if v.IsSensitive() {
				delete(config, k)
			} else {
				followCfg, ok := config[k].(map[string]any)
				if !ok {
					continue
				}

				config[k] = clearSensitiveValuesFromRead(v.NestedObject.Attributes, followCfg)
			}
		case schema.MapNestedAttribute:
			if v.IsSensitive() {
				delete(config, k)
			} else {
				followCfg, ok := config[k].(map[string]any)
				if !ok {
					continue
				}

				config[k] = clearSensitiveValuesFromRead(v.NestedObject.Attributes, followCfg)
			}
		case schema.SetNestedAttribute:
			if v.IsSensitive() {
				delete(config, k)
			} else {
				followCfg, ok := config[k].(map[string]any)
				if !ok {
					continue
				}

				config[k] = clearSensitiveValuesFromRead(v.NestedObject.Attributes, followCfg)
			}
		case schema.SingleNestedAttribute:
			if v.IsSensitive() {
				delete(config, k)
			} else {
				followCfg, ok := config[k].(map[string]any)
				if !ok {
					continue
				}

				config[k] = clearSensitiveValuesFromRead(v.Attributes, followCfg)
			}
		default:
			if v.IsSensitive() {
				delete(config, k)
			}
		}
	}
	return config
}

func getConfigAttributes(s schema.Schema) (map[string]schema.Attribute, bool) {
	attrsRaw, ok := s.Attributes["configuration"]
	if !ok {
		return nil, false
	}

	attrs, ok := attrsRaw.(schema.SingleNestedAttribute)
	if !ok {
		return nil, false
	}

	return attrs.Attributes, true
}

func handleSensitiveValues(ctx context.Context, attrs map[string]schema.Attribute, config map[string]any, priorState map[string]attr.Value) map[string]any {
	for k, v := range config {
		attr := attrs[k]

		switch v := attr.(type) {
		case schema.ListNestedAttribute:
			if v.IsSensitive() {
				delete(config, k)
			} else {
				nestedPstate, ok := priorState[k].(types.Object)
				if !ok {
					log.Printf("prior state for %s is not an object", k)
					continue
				}
				config[k] = handleSensitiveValues(ctx, v.NestedObject.Attributes, config[k].(map[string]any), nestedPstate.Attributes())
			}
			continue
		case schema.MapNestedAttribute:
			if v.IsSensitive() {
				delete(config, k)
			} else {
				nestedPstate, ok := priorState[k].(types.Object)
				if !ok {
					log.Printf("prior state for %s is not an object", k)
					continue
				}

				config[k] = handleSensitiveValues(ctx, v.NestedObject.Attributes, config[k].(map[string]any), nestedPstate.Attributes())
			}
			continue
		case schema.SetNestedAttribute:
			if v.IsSensitive() {
				delete(config, k)
			} else {
				nestedPstate, ok := priorState[k].(types.Object)
				if !ok {
					log.Printf("prior state for %s is not an object", k)
					continue
				}
				config[k] = handleSensitiveValues(ctx, v.NestedObject.Attributes, config[k].(map[string]any), nestedPstate.Attributes())
			}
			continue
		case schema.SingleNestedAttribute:
			if v.IsSensitive() {
				delete(config, k)
			} else {
				nestedPstate, ok := priorState[k].(types.Object)
				if !ok {
					log.Printf("prior state for %s is not an object", k)
					continue
				}
				config[k] = handleSensitiveValues(ctx, v.Attributes, config[k].(map[string]any), nestedPstate.Attributes())
			}
			continue
		}

		if attr.IsSensitive() {
			// if sensitive, see if the value equals the prior state, if it does clear it
			log.Printf("checking sensitive value for %s", k)
			compVal, err := attrValue(ctx, priorState[k])
			if err != nil {
				log.Printf("error getting prior state for %s: %s", k, err)
				continue
				// TODO: maybe return err?
			}

			if v == compVal {
				delete(config, k)
			}
		}
	}

	return config
}
