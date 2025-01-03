package connections

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestAttrValue(t *testing.T) {
	tests := map[string]struct {
		val       attr.Value
		expected  interface{}
		expectErr bool
	}{
		"string": {
			val:      types.StringValue("hello"),
			expected: "hello",
		},
		"object": {
			val: types.ObjectValueMust(
				map[string]attr.Type{
					"name": types.StringType,
					"age":  types.Int64Type,
				},
				map[string]attr.Value{
					"name": types.StringValue("Alice"),
					"age":  types.Int64Value(42),
				},
			),
			expected: map[string]interface{}{
				"name": "Alice",
				"age":  int64(42),
			},
		},
		"sets": {
			val: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("foo"),
				types.StringValue("bar"),
			}),
			expected: []interface{}{"foo", "bar"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := attrValue(context.Background(), test.val)
			if test.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, test.expected, actual)
			}
		})
	}
}
