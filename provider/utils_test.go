package provider

import "testing"

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
			actual := ValidNamer(tc.input)
			if actual != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}
