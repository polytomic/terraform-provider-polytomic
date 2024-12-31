package provider

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	legalCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
)

// A name must start with a letter or underscore and
// may contain only letters, digits, underscores, and dashes.
// e.g 100_users -> _100_users
func ValidName(s string) string {
	if len(s) == 0 {
		return "_"
	}

	// if string is not a letter or underscore, prepend underscore
	if !unicode.IsLetter(rune(s[0])) && s[0] != '_' {
		s = "_" + s
	}

	// replace illegal characters with underscore
	for i, v := range []byte(s) {
		if !strings.Contains(legalCharacters, string(v)) {
			s = s[:i] + "_" + s[i+1:]
		}
		if unicode.IsLower(rune(v)) && i < len(s)-1 && unicode.IsUpper(rune(s[i+1])) {
			s = s[:i+1] + "_" + strings.ToLower(s[i+1:])
		}
	}

	return s
}

func ToSnakeCase(s string) string {
	s = strings.TrimSpace(s)
	n := strings.Builder{}
	n.Grow(len(s) + 2) // nominal 2 bytes of extra space for inserted delimiters
	for i, v := range []byte(s) {
		vIsCap := v >= 'A' && v <= 'Z'
		vIsLow := v >= 'a' && v <= 'z'
		if vIsCap {
			v += 'a'
			v -= 'A'
		}

		if i+1 < len(s) {
			next := s[i+1]
			vIsNum := v >= '0' && v <= '9'
			nextIsCap := next >= 'A' && next <= 'Z'
			nextIsLow := next >= 'a' && next <= 'z'
			nextIsNum := next >= '0' && next <= '9'
			// add underscore if next letter case type is changed
			if (vIsCap && (nextIsLow)) || (vIsLow && (nextIsCap || nextIsNum)) || (vIsNum && (nextIsCap || nextIsLow)) {
				if vIsCap && nextIsLow {
					if prevIsCap := i > 0 && s[i-1] >= 'A' && s[i-1] <= 'Z'; prevIsCap {
						n.WriteByte('_')
					}
				}
				n.WriteByte(v)
				if vIsLow || vIsNum || nextIsNum {
					n.WriteByte('_')
				}
				continue

			}
		}

		if unicode.IsNumber(rune(v)) || unicode.IsLetter(rune(v)) {
			n.WriteByte(v)
		} else if n.Len() > 0 {
			n.WriteByte('_')
		}
	}

	return n.String()
}

func stringy(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case bool:
		return fmt.Sprintf("%t", t)
	default:
		panic(fmt.Sprintf("unsupported type %T", t))
	}
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

func attrValueString(v any) string {
	if s, ok := v.(types.String); ok {
		return s.ValueString()
	}
	return ""
}

func attrValueInt(v any) int {
	if s, ok := v.(types.Int64); ok {
		return int(s.ValueInt64())
	}
	return 0
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
