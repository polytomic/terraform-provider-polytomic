package importer

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
)

// convert configuration to map[string]cty.Value
func typeConverter(value any) cty.Value {
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
				vals = append(vals, typeConverter(v))
			}
			config[k] = cty.ListVal(vals)
		default:
			fmt.Printf("Unknown type for %s: %T\n", k, v)
			continue
		}
	}
	return cty.ObjectVal(config)
}
