package importer

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/polytomic/polytomic-go"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/slices"
)

// ReplaceRefs takes in an array of bytes and a map of references
// It will parse the array of bytes and replace any reference byte sequences
// with the value from the map.
func ReplaceRefs(b []byte, refs map[string]string) []byte {
	s := string(b)
	for k, v := range refs {
		// Replace all instances of the reference and the enclosing quotes
		re := regexp.MustCompile(`"` + k + `"`)
		s = re.ReplaceAllString(s, v)
	}
	return []byte(s)
}

func unquoteVariableRef(b []byte) []byte {
	s := string(b)
	re := regexp.MustCompile(`\"(var\..*)\"`)
	s = re.ReplaceAllString(s, "$1")
	return []byte(s)
}

// convert arbitrary values to cty.Value
func typeConverter(value any) cty.Value {
	switch value := value.(type) {
	case map[string]any:
		config := make(map[string]cty.Value)
		for k, v := range value {
			if v == nil {
				continue
			}
			switch v := v.(type) {
			case *string:
				if v == nil {
					continue
				} else if *v == "" {
					continue
				}
				config[k] = cty.StringVal(*v)
			case string:
				if v == "" {
					continue
				}
				config[k] = cty.StringVal(v)
			case int:
				config[k] = cty.NumberIntVal(int64(v))
			case *int:
				if v == nil {
					continue
				}
				config[k] = cty.NumberIntVal(int64(*v))
			case float64:
				config[k] = cty.NumberFloatVal(v)
			case *bool:
				if v == nil {
					continue
				}
				config[k] = cty.BoolVal(*v)
			case bool:
				config[k] = cty.BoolVal(v)
			case map[string]any:
				config[k] = typeConverter(v)
			case map[string]string:
				config[k] = typeConverter(v)
			case []string:
				if len(v) == 0 {
					continue
				}
				vals := make([]cty.Value, 0)
				for _, v := range v {
					vals = append(vals, cty.StringVal(v))
				}
				if len(vals) == 0 {
					continue
				}
				config[k] = cty.ListVal(vals)
			case []any:
				if len(v) == 0 {
					continue
				}
				vals := make([]cty.Value, 0)
				for _, v := range v {
					switch v := v.(type) {
					case map[string]any:
						vals = append(vals, typeConverter(v))
					case string:
						vals = append(vals, cty.StringVal(v))
					case int:
						vals = append(vals, cty.NumberIntVal(int64(v)))
					case float64:
						vals = append(vals, cty.NumberFloatVal(v))
					case bool:
						vals = append(vals, cty.BoolVal(v))
					default:
						fmt.Printf("Unknown type for %s: %T in array\n", k, v)
						continue
					}
				}
				if len(vals) == 0 {
					continue
				}
				config[k] = cty.ListVal(vals)
			case *polytomic.RunAfter:
				if v == nil {
					continue
				}
				config[k] = typeConverter(v)
			case *polytomic.Source:
				if v == nil {
					continue
				}
				config[k] = typeConverter(v)
			case polytomic.ScheduleFrequency:
				config[k] = cty.StringVal(string(v))
			default:
				fmt.Printf("Unknown type for %s: %T in map\n", k, v)
				continue
			}
		}
		return cty.ObjectVal(config)
	case []map[string]any:
		vals := make([]cty.Value, 0)
		for _, v := range value {
			vals = append(vals, typeConverter(v))
		}
		if len(vals) == 0 {
			return cty.EmptyTupleVal
		}
		return cty.TupleVal(vals)
	case []string:
		vals := make([]cty.Value, 0)
		for _, v := range value {
			vals = append(vals, cty.StringVal(v))
		}
		if len(vals) == 0 {
			return cty.ListValEmpty(cty.String)
		}
		return cty.ListVal(vals)
	case []any:
		if len(value) == 0 {
			return cty.NilVal
		}
		vals := make([]cty.Value, 0)
		for _, v := range value {
			switch v := v.(type) {
			case map[string]any:
				vals = append(vals, typeConverter(v))
			case string:
				vals = append(vals, cty.StringVal(v))
			case int:
				vals = append(vals, cty.NumberIntVal(int64(v)))
			case float64:
				vals = append(vals, cty.NumberFloatVal(v))
			case bool:
				vals = append(vals, cty.BoolVal(v))
			default:
				fmt.Printf("Unknown type for %s: %T in array\n", value, v)
				continue
			}
		}
		return cty.ListVal(vals)
	case map[string]*string:
		config := make(map[string]cty.Value)
		for k, v := range value {
			if v == nil {
				continue
			}
			if *v != "" {
				config[k] = cty.StringVal(*v)
			}
		}
		return cty.ObjectVal(config)
	case map[string]string:
		config := make(map[string]cty.Value)
		for k, v := range value {
			if v != "" {
				config[k] = cty.StringVal(v)
			}
		}
		return cty.ObjectVal(config)
	case *string:
		if value == nil {
			return cty.NilVal
		}
		return cty.StringVal(*value)
	case string:
		if value == "" {
			return cty.NilVal
		}
		return cty.StringVal(value)
	case nil:
		return cty.NilVal
	case bool:
		return cty.BoolVal(value)
	case float64:
		return cty.NumberFloatVal(value)
	case *polytomic.Source:
		if value == nil {
			return cty.NilVal
		}
		config := map[string]cty.Value{
			"field":    cty.StringVal(value.Field),
			"model_id": cty.StringVal(value.ModelId),
		}
		return cty.ObjectVal(config)
	default:
		fmt.Printf("Unknown type: %T\n", value)
		return cty.NilVal
	}
}

// wrapJSONEncode wraps the given attribute names within the given map[string]any
// or []map[string]any with a jsonencode function.
func wrapJSONEncode(v interface{}, wrapped ...string) hclwrite.Tokens {
	var tokens hclwrite.Tokens

	if len(wrapped) == 0 {
		// Assume the whole thing is wrapped
		tokens = append(tokens, hclwrite.Tokens{{Bytes: []byte("jsonencode(")}}...)
		tokens = append(tokens, hclwrite.TokensForValue(typeConverter(v))...)
		tokens = append(tokens, hclwrite.Tokens{{Bytes: []byte(")")}}...)
		return tokens
	}

	switch v := v.(type) {
	case map[string]any:
		tokens = append(tokens, jsonEncodeMap(v, wrapped...)...)
	case []map[string]any:
		tokens = append(tokens, hclwrite.Tokens{{Bytes: []byte("[")}}...)
		for _, v := range v {
			tokens = append(tokens, jsonEncodeMap(v, wrapped...)...)
			tokens = append(tokens, hclwrite.Tokens{{Bytes: []byte(",")}}...)
		}
		tokens = append(tokens, hclwrite.Tokens{{Bytes: []byte("]")}}...)
	default:
		fmt.Printf("Unknown type: %T in jsonencode wrapper\n", v)
	}

	return tokens
}

// jsonEncodeMap wraps the given attribute names with a jsonencode function.
func jsonEncodeMap(m map[string]any, wrapped ...string) hclwrite.Tokens {
	var tokens hclwrite.Tokens
	tokens = append(tokens, &hclwrite.Token{Bytes: []byte("{\n")})
	for _, k := range sortedKeys(m) {
		v := m[k]
		value := typeConverter(v)
		if value == cty.NilVal {
			continue
		}
		if slices.Contains(wrapped, k) {
			tokens = append(tokens, &hclwrite.Token{Bytes: []byte(k)})
			tokens = append(tokens, &hclwrite.Token{Bytes: []byte("=")})
			tokens = append(tokens, &hclwrite.Token{Bytes: []byte("jsonencode(")})
			tokens = append(tokens, hclwrite.TokensForValue(typeConverter(v))...)
			tokens = append(tokens, &hclwrite.Token{Bytes: []byte(")")})
			tokens = append(tokens, &hclwrite.Token{Bytes: []byte("\n")})

		} else {
			tokens = append(tokens, &hclwrite.Token{Bytes: []byte(k)})
			tokens = append(tokens, &hclwrite.Token{Bytes: []byte("=")})
			tokens = append(tokens, hclwrite.TokensForValue(typeConverter(v))...)
			tokens = append(tokens, &hclwrite.Token{Bytes: []byte("\n")})
		}
	}
	tokens = append(tokens, &hclwrite.Token{Bytes: []byte("}")})

	return tokens
}
