package importer

import (
	"encoding/json"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/slices"
)

// Encode JSON-valued attributes directly, preserving empty strings, nulls,
// empty collections, and mixed arrays that the structural converter omits.
func wrapJSONAttributes(value any, wrapped ...string) (hclwrite.Tokens, error) {
	switch value := value.(type) {
	case map[string]any:
		file := hclwrite.NewEmptyFile()
		for _, key := range sortedKeys(value) {
			if slices.Contains(wrapped, key) {
				data, err := json.Marshal(value[key])
				if err != nil {
					return nil, err
				}
				tokens, err := jsonEncodeTokens(data)
				if err != nil {
					return nil, err
				}
				file.Body().SetAttributeRaw(key, tokens)
			} else if text, ok := value[key].(string); ok {
				file.Body().SetAttributeValue(key, cty.StringVal(text))
			} else if text, ok := value[key].(*string); ok && text != nil {
				file.Body().SetAttributeValue(key, cty.StringVal(*text))
			} else {
				converted := typeConverter(map[string]any{key: value[key]})
				for key, attribute := range converted.AsValueMap() {
					file.Body().SetAttributeValue(key, attribute)
				}
			}
		}
		tokens := hclwrite.Tokens{{Bytes: []byte("{\n")}}
		tokens = append(tokens, file.Body().BuildTokens(nil)...)
		return append(tokens, &hclwrite.Token{Bytes: []byte("}")}), nil
	case []map[string]any:
		tokens := hclwrite.Tokens{{Bytes: []byte("[")}}
		for _, item := range value {
			encoded, err := wrapJSONAttributes(item, wrapped...)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, encoded...)
			tokens = append(tokens, &hclwrite.Token{Bytes: []byte(",")})
		}
		return append(tokens, &hclwrite.Token{Bytes: []byte("]")}), nil
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		return jsonEncodeTokens(data)
	}
}
