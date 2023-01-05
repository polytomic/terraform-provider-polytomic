package importer

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

// convert arbitrary values to cty.Value
func typeConverter(value any) cty.Value {
	switch value.(type) {
	case map[string]any:
		config := make(map[string]cty.Value)
		for k, v := range value.(map[string]any) {
			switch v.(type) {
			case string:
				config[k] = cty.StringVal(v.(string))
			case int:
				config[k] = cty.NumberIntVal(int64(v.(int)))
			case float64:
				config[k] = cty.NumberFloatVal(v.(float64))
			case bool:
				config[k] = cty.BoolVal(v.(bool))
			case map[string]any:
				config[k] = typeConverter(v)
			case []any:
				if len(v.([]any)) == 0 {
					continue
				}
				vals := make([]cty.Value, 0)
				for _, v := range v.([]any) {
					switch v.(type) {
					case map[string]any:
						vals = append(vals, typeConverter(v))
					case string:
						vals = append(vals, cty.StringVal(v.(string)))
					case int:
						vals = append(vals, cty.NumberIntVal(int64(v.(int))))
					case float64:
						vals = append(vals, cty.NumberFloatVal(v.(float64)))
					case bool:
						vals = append(vals, cty.BoolVal(v.(bool)))
					default:
						fmt.Printf("Unknown type for %s: %T\n", k, v)
						continue
					}
				}
				if len(vals) == 0 {
					continue
				}
				config[k] = cty.ListVal(vals)
			default:
				fmt.Printf("Unknown type for %s: %T\n", k, v)
				continue
			}
		}
		return cty.ObjectVal(config)
	case []map[string]any:
		vals := make([]cty.Value, 0)
		for _, v := range value.([]map[string]any) {
			vals = append(vals, typeConverter(v))
		}
		if len(vals) == 0 {
			return cty.NilVal
		}
		return cty.ListVal(vals)
	case []string:
		vals := make([]cty.Value, 0)
		for _, v := range value.([]string) {
			vals = append(vals, cty.StringVal(v))
		}
		return cty.ListVal(vals)
	default:
		fmt.Printf("Unknown type: %T\n", value)
		return cty.NilVal
	}

	return cty.NilVal
}
