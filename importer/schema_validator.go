package importer

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// SchemaValidator validates that field mappings match a provider resource schema
type SchemaValidator struct {
	resourceSchema schema.Schema
}

// NewSchemaValidator creates a validator from a provider resource
func NewSchemaValidator(ctx context.Context, res resource.Resource) (*SchemaValidator, error) {
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	res.Schema(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		return nil, fmt.Errorf("failed to get schema: %v", resp.Diagnostics.Errors())
	}

	return &SchemaValidator{
		resourceSchema: resp.Schema,
	}, nil
}

// ValidateMapping validates that a field mapping matches the schema structure
// mapping is a nested map[string]interface{} representing the HCL structure
// Returns an error if any field paths don't exist in the schema
func (v *SchemaValidator) ValidateMapping(mapping map[string]interface{}) error {
	return v.validateLevel(mapping, v.resourceSchema.Attributes, "")
}

func (v *SchemaValidator) validateLevel(mapping map[string]interface{}, attrs map[string]schema.Attribute, path string) error {
	for fieldName, value := range mapping {
		currentPath := fieldName
		if path != "" {
			currentPath = path + "." + fieldName
		}

		// Check if field exists in schema
		attr, exists := attrs[fieldName]
		if !exists {
			// Skip computed-only fields that won't be in user input
			if isComputedOnlyField(fieldName) {
				continue
			}
			return fmt.Errorf("field '%s' not found in schema%s", currentPath, v.suggestAlternative(fieldName, attrs))
		}

		// If value is a nested map, recursively validate
		if nestedMap, ok := value.(map[string]interface{}); ok {
			nestedAttrs := v.getNestedAttributes(attr)
			if nestedAttrs == nil {
				return fmt.Errorf("field '%s' is not a nested object in schema", currentPath)
			}
			if err := v.validateLevel(nestedMap, nestedAttrs, currentPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *SchemaValidator) getNestedAttributes(attr schema.Attribute) map[string]schema.Attribute {
	switch a := attr.(type) {
	case schema.SingleNestedAttribute:
		return a.Attributes
	case schema.ListNestedAttribute:
		return a.NestedObject.Attributes
	case schema.SetNestedAttribute:
		return a.NestedObject.Attributes
	case schema.MapNestedAttribute:
		return a.NestedObject.Attributes
	default:
		return nil
	}
}

func (v *SchemaValidator) suggestAlternative(fieldName string, attrs map[string]schema.Attribute) string {
	// Convert old field names to new ones
	suggestions := map[string]string{
		"source_connection_id": "source.connection_id",
		"dest_connection_id":   "destination.connection_id",
		"source_configuration": "source.configuration",
		"dest_configuration":   "destination.configuration",
		"discover":             "automatically_add_new_fields or automatically_add_new_objects",
	}

	if suggestion, ok := suggestions[fieldName]; ok {
		return fmt.Sprintf(", did you mean '%s'?", suggestion)
	}

	// Find similar field names
	fieldLower := strings.ToLower(fieldName)
	for schemaField := range attrs {
		if strings.Contains(strings.ToLower(schemaField), fieldLower) ||
			strings.Contains(fieldLower, strings.ToLower(schemaField)) {
			return fmt.Sprintf(", did you mean '%s'?", schemaField)
		}
	}

	return ""
}

func isComputedOnlyField(fieldName string) bool {
	computedFields := map[string]bool{
		"id":           true,
		"created_at":   true,
		"updated_at":   true,
		"version":      true,
		"organization": true, // Often computed with default
	}
	return computedFields[fieldName]
}
