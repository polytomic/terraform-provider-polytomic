package connections

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
func TestHandleSensitiveValues(t *testing.T) {
	tests := map[string]struct {
		attrs    map[string]schema.Attribute
		config   map[string]any
		data     connectionData
		expected map[string]any
	}{
		"no sensitive values": {
			attrs: map[string]schema.Attribute{
				"name": schema.StringAttribute{},
				"age":  schema.Int64Attribute{},
			},
			config: map[string]any{
				"name": "Alice",
				"age":  int64(42),
			},
			data: connectionData{
				Configuration: types.ObjectValueMust(
					map[string]attr.Type{
						"name": types.StringType,
						"age":  types.Int64Type,
					},
					map[string]attr.Value{
						"name": types.StringValue("Alice"),
						"age":  types.Int64Value(42),
					},
				),
			},
			expected: map[string]any{
				"name": "Alice",
				"age":  int64(42),
			},
		},
		"sensitive values": {
			attrs: map[string]schema.Attribute{
				"password": schema.StringAttribute{Sensitive: true},
				"token":    schema.StringAttribute{Sensitive: true},
			},
			config: map[string]any{
				"password": "secret",
				"token":    "token123",
			},
			data: connectionData{
				Configuration: types.ObjectValueMust(
					map[string]attr.Type{
						"password": types.StringType,
						"token":    types.StringType,
					},
					map[string]attr.Value{
						"password": types.StringValue("secret"),
						"token":    types.StringValue("token123"),
					},
				),
			},
			expected: map[string]any{},
		},
		"sensitive values updated": {
			attrs: map[string]schema.Attribute{
				"password": schema.StringAttribute{Sensitive: true},
			},
			config: map[string]any{
				"password": "secret123",
			},
			data: connectionData{
				Configuration: types.ObjectValueMust(
					map[string]attr.Type{
						"password": types.StringType,
					},
					map[string]attr.Value{
						"password": types.StringValue("secret"),
					},
				),
			},
			expected: map[string]any{
				"password": "secret123",
			},
		},
		"mixed sensitive and non-sensitive values": {
			attrs: map[string]schema.Attribute{
				"name":     schema.StringAttribute{},
				"password": schema.StringAttribute{Sensitive: true},
			},
			config: map[string]any{
				"name":     "Alice",
				"password": "secret",
			},
			data: connectionData{
				Configuration: types.ObjectValueMust(
					map[string]attr.Type{
						"name":     types.StringType,
						"password": types.StringType,
					},
					map[string]attr.Value{
						"name":     types.StringValue("Alice"),
						"password": types.StringValue("secret"),
					},
				),
			},
			expected: map[string]any{
				"name": "Alice",
			},
		},
		"nested sensitive values": {
			attrs: map[string]schema.Attribute{
				"nestedSingle": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"password": schema.StringAttribute{Sensitive: true},
						"name":     schema.StringAttribute{},
					},
				},
				"nestedMap": schema.MapNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"password": schema.StringAttribute{Sensitive: true},
							"name":     schema.StringAttribute{},
						},
					},
				},
				"nestedSet": schema.SetNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"password": schema.StringAttribute{Sensitive: true},
							"name":     schema.StringAttribute{},
						},
					},
				},
				"nestedList": schema.ListNestedAttribute{
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"password": schema.StringAttribute{Sensitive: true},
							"name":     schema.StringAttribute{},
						},
					},
				},
			},
			config: map[string]any{
				"nestedSingle": map[string]any{
					"name":     "Alice",
					"password": "secret",
				},
				"nestedMap": map[string]any{
					"name":     "Alice",
					"password": "secret",
				},
				"nestedSet": map[string]any{
					"name":     "Alice",
					"password": "secret",
				},
				"nestedList": map[string]any{
					"name":     "Alice",
					"password": "secret",
				},
			},
			data: connectionData{
				Configuration: types.ObjectValueMust(
					map[string]attr.Type{
						"nestedSingle": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name":     types.StringType,
								"password": types.StringType,
							},
						},
						"nestedMap": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name":     types.StringType,
								"password": types.StringType,
							},
						},
						"nestedSet": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name":     types.StringType,
								"password": types.StringType,
							},
						},
						"nestedList": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name":     types.StringType,
								"password": types.StringType,
							},
						},
					},
					map[string]attr.Value{
						"nestedSingle": types.ObjectValueMust(
							map[string]attr.Type{
								"name":     types.StringType,
								"password": types.StringType,
							},
							map[string]attr.Value{
								"name":     types.StringValue("Alice"),
								"password": types.StringValue("secret"),
							},
						),
						"nestedMap": types.ObjectValueMust(
							map[string]attr.Type{
								"name":     types.StringType,
								"password": types.StringType,
							},
							map[string]attr.Value{
								"name":     types.StringValue("Alice"),
								"password": types.StringValue("secret"),
							},
						),
						"nestedSet": types.ObjectValueMust(
							map[string]attr.Type{
								"name":     types.StringType,
								"password": types.StringType,
							},
							map[string]attr.Value{
								"name":     types.StringValue("Alice"),
								"password": types.StringValue("secret"),
							},
						),
						"nestedList": types.ObjectValueMust(
							map[string]attr.Type{
								"name":     types.StringType,
								"password": types.StringType,
							},
							map[string]attr.Value{
								"name":     types.StringValue("Alice"),
								"password": types.StringValue("secret"),
							},
						),
					},
				),
			},
			expected: map[string]any{
				"nestedSingle": map[string]any{
					"name": "Alice",
				},
				"nestedMap": map[string]any{
					"name": "Alice",
				},
				"nestedSet": map[string]any{
					"name": "Alice",
				},
				"nestedList": map[string]any{
					"name": "Alice",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := handleSensitiveValues(context.Background(), test.attrs, test.config, test.data.Configuration.Attributes())
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
func TestResetSensitiveValues(t *testing.T) {
	tests := map[string]struct {
		attrs    map[string]schema.Attribute
		state    map[string]any
		read     map[string]any
		expected map[string]any
	}{
		"no sensitive values": {
			attrs: map[string]schema.Attribute{
				"name": schema.StringAttribute{},
				"age":  schema.Int64Attribute{},
			},
			state: map[string]any{
				"name": "Alice",
				"age":  int64(42),
			},
			read: map[string]any{
				"name": "Alice",
				"age":  int64(42),
			},
			expected: map[string]any{
				"name": "Alice",
				"age":  int64(42),
			},
		},
		"sensitive values": {
			attrs: map[string]schema.Attribute{
				"password": schema.StringAttribute{Sensitive: true},
				"token":    schema.StringAttribute{Sensitive: true},
			},
			state: map[string]any{
				"password": "secret",
				"token":    "token123",
			},
			read: map[string]any{
				"password": "new_secretdasdfasdf",
				"token":    "new_token123adsfasdfa",
			},
			expected: map[string]any{
				"password": "secret",
				"token":    "token123",
			},
		},
		"mixed sensitive and non-sensitive values": {
			attrs: map[string]schema.Attribute{
				"name":     schema.StringAttribute{},
				"password": schema.StringAttribute{Sensitive: true},
			},
			state: map[string]any{
				"name":     "Alice",
				"password": "secret",
			},
			read: map[string]any{
				"name":     "Alice",
				"password": "new_secret",
			},
			expected: map[string]any{
				"name":     "Alice",
				"password": "secret",
			},
		},
		"nested sensitive values": {
			attrs: map[string]schema.Attribute{
				"nestedSingle": schema.SingleNestedAttribute{
					Attributes: map[string]schema.Attribute{
						"password": schema.StringAttribute{Sensitive: true},
						"name":     schema.StringAttribute{},
					},
				},
			},
			state: map[string]any{
				"nestedSingle": map[string]any{
					"name":     "Alice",
					"password": "secret",
				},
			},
			read: map[string]any{
				"nestedSingle": map[string]any{
					"name":     "Alice",
					"password": "new_secret",
				},
			},
			expected: map[string]any{
				"nestedSingle": map[string]any{
					"name":     "Alice",
					"password": "secret",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := resetSensitiveValues(test.attrs, test.state, test.read)
			assert.EqualValues(t, test.expected, actual)
		})
	}
}
