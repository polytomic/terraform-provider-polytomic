package importer

import (
	"context"
	"strings"
	"testing"

	"github.com/polytomic/terraform-provider-polytomic/provider"
)

// TestSchemaValidatorMapAttribute verifies that a map attribute accepts
// arbitrary keys, while unknown fields of nested objects are still rejected.
func TestSchemaValidatorMapAttribute(t *testing.T) {
	v, err := NewSchemaValidator(context.Background(), provider.ConnectionsMap["awsathena"])
	if err != nil {
		t.Fatal(err)
	}

	for _, tags := range []map[string]any{{}, {"team": "data"}} {
		err := v.ValidateMapping(map[string]any{
			"name":          "Athena",
			"configuration": map[string]any{"tags": tags},
		})
		if err != nil {
			t.Errorf("tags %v: %v", tags, err)
		}
	}

	err = v.ValidateMapping(map[string]any{
		"configuration": map[string]any{"no_such_field": "x"},
	})
	if err == nil || !strings.Contains(err.Error(), "configuration.no_such_field") {
		t.Errorf("want an error naming configuration.no_such_field, got %v", err)
	}
}
