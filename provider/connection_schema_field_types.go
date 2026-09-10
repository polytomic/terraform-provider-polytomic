package provider

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/polytomic/polytomic-go/v25"
)

// fieldTypeNames are the values polytomic_connection_schema_field's type
// attribute accepts: the basic field types, then the detailed ones.
var fieldTypeNames = []string{
	"string", "number", "boolean", "datetime", "array", "object", "binary",
	"smallint", "int", "bigint", "single", "double", "decimal", "date", "time", "datetime_tz",
}

// basicFieldTypes pairs each basic field type with the definition the API
// stores for it by default. Sending it explicitly replaces any detailed
// definition, such as one inherited from a detected field.
var basicFieldTypes = map[string]any{
	"string":   "string",
	"number":   "number",
	"boolean":  "boolean",
	"datetime": "datetime",
	"binary":   "binary",
	"array":    []any{"jsonarray"},
	"object":   []any{"json"},
}

// detailedFieldTypes maps each detailed type name to its basic field type.
// Every name except decimal is also its own definition.
var detailedFieldTypes = map[string]string{
	"smallint":    "number",
	"int":         "number",
	"bigint":      "number",
	"single":      "number",
	"double":      "number",
	"decimal":     "number",
	"date":        "datetime",
	"time":        "datetime",
	"datetime_tz": "datetime",
}

// The SDK cannot decode a named definition, so named definitions are built
// with its constructors.
var typesDefinitionByName = map[string]func() *polytomic.TypesDefinition{
	"binary":      polytomic.NewTypesDefinitionWithBinaryStringLiteral,
	"boolean":     polytomic.NewTypesDefinitionWithBooleanStringLiteral,
	"date":        polytomic.NewTypesDefinitionWithDateStringLiteral,
	"datetime":    polytomic.NewTypesDefinitionWithDatetimeStringLiteral,
	"datetime_tz": polytomic.NewTypesDefinitionWithDatetime_tzStringLiteral,
	"time":        polytomic.NewTypesDefinitionWithTimeStringLiteral,
	"number":      polytomic.NewTypesDefinitionWithNumberStringLiteral,
	"string":      polytomic.NewTypesDefinitionWithFieldStringStringLiteral,
	"smallint":    polytomic.NewTypesDefinitionWithSmallintStringLiteral,
	"int":         polytomic.NewTypesDefinitionWithIntStringLiteral,
	"bigint":      polytomic.NewTypesDefinitionWithBigintStringLiteral,
	"single":      polytomic.NewTypesDefinitionWithSingleStringLiteral,
	"double":      polytomic.NewTypesDefinitionWithDoubleStringLiteral,
}

// fieldTypeSpec returns the API field type and definition for a type name.
func fieldTypeSpec(name string, precision, scale int64) (basic string, spec any, err error) {
	if name == "decimal" {
		return "number", []any{"decimal", map[string]any{"precision": precision, "scale": scale}}, nil
	}
	if spec, ok := basicFieldTypes[name]; ok {
		return name, spec, nil
	}
	if basic, ok := detailedFieldTypes[name]; ok {
		return basic, name, nil
	}
	return "", nil, fmt.Errorf("unknown field type %q", name)
}

// fieldTypeFromSpec is the inverse of fieldTypeSpec for a definition decoded
// from JSON. ok is false when no type name describes the definition.
func fieldTypeFromSpec(spec any) (name string, precision, scale *int64, ok bool) {
	switch v := spec.(type) {
	case string:
		if s, isBasic := basicFieldTypes[v].(string); isBasic && s == v {
			return v, nil, nil, true
		}
		if _, isDetailed := detailedFieldTypes[v]; isDetailed && v != "decimal" {
			return v, nil, nil, true
		}
	case []any:
		switch {
		case len(v) == 1 && v[0] == "jsonarray":
			return "array", nil, nil, true
		case len(v) == 1 && v[0] == "json":
			return "object", nil, nil, true
		case len(v) == 2 && v[0] == "decimal":
			details, _ := v[1].(map[string]any)
			p, pok := jsonInt(details["precision"])
			s, sok := jsonInt(details["scale"])
			if pok && sok {
				return "decimal", &p, &s, true
			}
		}
	}
	return "", nil, nil, false
}

// specBasicType returns the basic field type a definition belongs to.
func specBasicType(spec any) (string, error) {
	switch v := spec.(type) {
	case string:
		if s, isBasic := basicFieldTypes[v].(string); isBasic && s == v {
			return v, nil
		}
		if basic, isDetailed := detailedFieldTypes[v]; isDetailed && v != "decimal" {
			return basic, nil
		}
		return "", fmt.Errorf("unknown type name %q", v)
	case []any:
		if len(v) > 0 {
			switch v[0] {
			case "array", "jsonarray":
				return "array", nil
			case "map", "object", "json":
				return "object", nil
			case "string":
				return "string", nil
			case "decimal":
				return "number", nil
			case "versioned":
				if len(v) == 2 {
					if versions, _ := v[1].([]any); len(versions) > 0 {
						return specBasicType(versions[0])
					}
				}
			}
		}
		return "", fmt.Errorf("unrecognized type_spec %v", v)
	}
	return "", errors.New("type_spec must be a type name or an array")
}

func newTypesDefinition(spec any) (*polytomic.TypesDefinition, error) {
	switch v := spec.(type) {
	case string:
		if ctor, ok := typesDefinitionByName[v]; ok {
			return ctor(), nil
		}
		return nil, fmt.Errorf("unknown type name %q", v)
	case []any:
		if len(v) > 0 {
			return &polytomic.TypesDefinition{UnknownList: v}, nil
		}
	}
	return nil, errors.New("type_spec must be a type name or a non-empty array")
}

// SchemaFieldTypeAttributes returns the polytomic_connection_schema_field
// attributes that describe a field's type: a type name, with precision and
// scale for decimal, or typeSpec as JSON when no type name describes it.
func SchemaFieldTypeAttributes(f *polytomic.SchemaField) (name string, precision, scale *int64, typeSpec string, err error) {
	if f.TypeSpec == nil || *f.TypeSpec == nil {
		if f.Type == nil {
			return "", nil, nil, "", nil
		}
		return string(*f.Type), nil, nil, "", nil
	}
	if name, precision, scale, ok := fieldTypeFromSpec(*f.TypeSpec); ok {
		return name, precision, scale, "", nil
	}
	buf, err := json.Marshal(*f.TypeSpec)
	if err != nil {
		return "", nil, nil, "", fmt.Errorf("encoding type_spec for field %s: %w", pointer.GetString(f.ID), err)
	}
	return "", nil, nil, string(buf), nil
}

func jsonInt(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), n == float64(int64(n))
	case int64:
		return n, true
	case int:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	}
	return 0, false
}
