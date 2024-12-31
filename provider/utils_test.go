package provider

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

func TestValidNamer(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "_",
		},
		{
			name:     "starts with number",
			input:    "100_users",
			expected: "_100_users",
		},
		{
			name:     "starts with underscore",
			input:    "_users",
			expected: "_users",
		},
		{
			name:     "starts with letter",
			input:    "users",
			expected: "users",
		},
		{
			name:     "contains illegal characters",
			input:    "users@",
			expected: "users_",
		},
		{
			name:     "camel case",
			input:    "camelCase",
			expected: "camel_case",
		},
		{
			name:     "snake case",
			input:    "snake_case",
			expected: "snake_case",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ValidName(tc.input)
			if actual != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}
