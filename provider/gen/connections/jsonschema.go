package connections

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// getJSONSchema returns a JSON schema from a Polytomic API representation.
func getJSONSchema(p *polytomic.JsonschemaSchema) (*jsonschema.Schema, error) {
	a := map[string]interface{}{}
	propJSON, _ := json.Marshal(p)
	err := json.Unmarshal(propJSON, &a)
	if err != nil {
		return nil, err
	}
	return unmarshalJSONSchema(a)
}

// unmarshalJSONSchema unmarshals a JSON schema from a map[string]interface{}.
// It recursively unmarshals the properties of the schema in order to populate
// any Extras that may be present.
func unmarshalJSONSchema(input map[string]interface{}) (*jsonschema.Schema, error) {
	a := jsonschema.Schema{}
	m := &mapstructure.Metadata{}
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			Result:           &a,
			Metadata:         m,
			WeaklyTypedInput: true,
			TagName:          "json",
			DecodeHook:       falseForNilSchema(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error creating mapstructure decoder: %w", err)
	}
	err = decoder.Decode(input)
	if err != nil {
		return nil, fmt.Errorf("error decoding schema: %w", err)
	}
	if len(m.Unused) > 0 {
		if a.Extras == nil {
			a.Extras = make(map[string]interface{}, len(m.Unused))
			for _, k := range m.Unused {
				a.Extras[k] = input[k]
			}
		}
	}
	if props, ok := input["properties"].(map[string]interface{}); ok {
		a.Properties = orderedmap.New[string, *jsonschema.Schema]()
		for _, k := range slices.Sorted(maps.Keys(props)) {
			v := props[k]
			if v == nil {
				continue
			}
			s, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			schema, err := unmarshalJSONSchema(s)
			if err != nil {
				return nil, fmt.Errorf("error decoding schema for %s: %w", k, err)
			}
			a.Properties.Set(k, schema)
		}
	}
	if rawItems, ok := input["items"]; ok {
		a.Items, err = unmarshalJSONSchema(rawItems.(map[string]interface{}))
		if err != nil {
			return nil, fmt.Errorf("error decoding items: %w", err)
		}
	}
	if rawAdditionalProps, ok := input["additionalProperties"]; ok {
		if additionalPropsMap, ok := rawAdditionalProps.(map[string]interface{}); ok {
			a.AdditionalProperties, err = unmarshalJSONSchema(additionalPropsMap)
			if err != nil {
				return nil, fmt.Errorf("error decoding additionalProperties: %w", err)
			}
		}
	}

	// Unmarshal conditional/composition keywords so that dependentSchemas,
	// oneOf, allOf, if/then/else are available for dependency extraction.
	if rawDepSchemas, ok := input["dependentSchemas"].(map[string]interface{}); ok {
		a.DependentSchemas = make(map[string]*jsonschema.Schema, len(rawDepSchemas))
		for k, v := range rawDepSchemas {
			if m, ok := v.(map[string]interface{}); ok {
				s, err := unmarshalJSONSchema(m)
				if err != nil {
					return nil, fmt.Errorf("error decoding dependentSchemas[%s]: %w", k, err)
				}
				a.DependentSchemas[k] = s
			}
		}
	}
	if rawOneOf, ok := input["oneOf"].([]interface{}); ok {
		a.OneOf = make([]*jsonschema.Schema, 0, len(rawOneOf))
		for i, v := range rawOneOf {
			if m, ok := v.(map[string]interface{}); ok {
				s, err := unmarshalJSONSchema(m)
				if err != nil {
					return nil, fmt.Errorf("error decoding oneOf[%d]: %w", i, err)
				}
				a.OneOf = append(a.OneOf, s)
			}
		}
	}
	if rawAllOf, ok := input["allOf"].([]interface{}); ok {
		a.AllOf = make([]*jsonschema.Schema, 0, len(rawAllOf))
		for i, v := range rawAllOf {
			if m, ok := v.(map[string]interface{}); ok {
				s, err := unmarshalJSONSchema(m)
				if err != nil {
					return nil, fmt.Errorf("error decoding allOf[%d]: %w", i, err)
				}
				a.AllOf = append(a.AllOf, s)
			}
		}
	}
	if rawIf, ok := input["if"].(map[string]interface{}); ok {
		a.If, err = unmarshalJSONSchema(rawIf)
		if err != nil {
			return nil, fmt.Errorf("error decoding if: %w", err)
		}
	}
	if rawThen, ok := input["then"].(map[string]interface{}); ok {
		a.Then, err = unmarshalJSONSchema(rawThen)
		if err != nil {
			return nil, fmt.Errorf("error decoding then: %w", err)
		}
	}
	if rawElse, ok := input["else"].(map[string]interface{}); ok {
		a.Else, err = unmarshalJSONSchema(rawElse)
		if err != nil {
			return nil, fmt.Errorf("error decoding else: %w", err)
		}
	}
	if rawContains, ok := input["contains"].(map[string]interface{}); ok {
		a.Contains, err = unmarshalJSONSchema(rawContains)
		if err != nil {
			return nil, fmt.Errorf("error decoding contains: %w", err)
		}
	}
	// Preserve const value if present (mapstructure may skip it).
	if rawConst, ok := input["const"]; ok {
		a.Const = rawConst
	}

	return &a, nil
}

func falseForNilSchema() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.Bool {
			return data, nil
		}
		if t != reflect.TypeOf(&jsonschema.Schema{}) {
			return data, nil
		}

		if data == false {
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected value: %V", data)
	}
}
