package connections

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttributesForJSONSchemaWithDependentSchemas(t *testing.T) {
	t.Run("oneOf dependencies", func(t *testing.T) {
		// Simulates a schema where auth_mode controls which fields are visible.
		props := jsonschema.NewProperties()
		props.Set("auth_mode", &jsonschema.Schema{
			Type:  "string",
			Title: "Authentication Method",
			Enum:  []interface{}{"access_key", "iam_role"},
		})
		props.Set("bucket", &jsonschema.Schema{
			Type:  "string",
			Title: "Bucket",
		})

		// Build oneOf branches for the dependent schema.
		akProps := jsonschema.NewProperties()
		akProps.Set("auth_mode", &jsonschema.Schema{Enum: []interface{}{"access_key"}})
		akProps.Set("access_key_id", &jsonschema.Schema{Type: "string", Title: "Access Key ID"})

		iamProps := jsonschema.NewProperties()
		iamProps.Set("auth_mode", &jsonschema.Schema{Enum: []interface{}{"iam_role"}})
		iamProps.Set("role_arn", &jsonschema.Schema{Type: "string", Title: "Role ARN"})

		schema := &jsonschema.Schema{
			Type:       "object",
			Properties: props,
			Required:   []string{"auth_mode", "bucket"},
			DependentSchemas: map[string]*jsonschema.Schema{
				"auth_mode": {
					OneOf: []*jsonschema.Schema{
						{Properties: akProps, Required: []string{"access_key_id"}},
						{Properties: iamProps, Required: []string{"role_arn"}},
					},
				},
			},
		}

		attrs, err := attributesForJSONSchema(schema)
		require.NoError(t, err)

		// Should have 4 attributes: auth_mode, bucket, access_key_id, role_arn.
		require.Len(t, attrs, 4)

		// Find the conditional attributes.
		var akAttr, iamAttr Attribute
		for _, a := range attrs {
			switch a.Name {
			case "access_key_id":
				akAttr = a
			case "role_arn":
				iamAttr = a
			}
		}

		// access_key_id should be conditional on auth_mode = "access_key".
		require.Len(t, akAttr.Conditions, 1)
		assert.Equal(t, "auth_mode", akAttr.Conditions[0].Field)
		assert.Equal(t, "access_key", akAttr.Conditions[0].Value)
		assert.True(t, akAttr.Conditions[0].Required)
		assert.False(t, akAttr.Required, "conditional attrs should be optional in terraform")
		assert.True(t, akAttr.Optional)
		assert.Contains(t, akAttr.Description, "Only applicable when")

		// role_arn should be conditional on auth_mode = "iam_role".
		require.Len(t, iamAttr.Conditions, 1)
		assert.Equal(t, "auth_mode", iamAttr.Conditions[0].Field)
		assert.Equal(t, "iam_role", iamAttr.Conditions[0].Value)
		assert.True(t, iamAttr.Conditions[0].Required)
		assert.Contains(t, iamAttr.Description, "Only applicable when")
	})

	t.Run("if-then dependencies", func(t *testing.T) {
		props := jsonschema.NewProperties()
		props.Set("ssh", &jsonschema.Schema{Type: "boolean", Title: "Use SSH"})

		// dependent schema: if ssh has a value, show ssh_host.
		ifProps := jsonschema.NewProperties()
		ifProps.Set("ssh", &jsonschema.Schema{Const: true})

		thenProps := jsonschema.NewProperties()
		thenProps.Set("ssh_host", &jsonschema.Schema{Type: "string", Title: "SSH Host"})

		schema := &jsonschema.Schema{
			Type:       "object",
			Properties: props,
			DependentSchemas: map[string]*jsonschema.Schema{
				"ssh": {
					If:   &jsonschema.Schema{Properties: ifProps},
					Then: &jsonschema.Schema{Properties: thenProps, Required: []string{"ssh_host"}},
				},
			},
		}

		attrs, err := attributesForJSONSchema(schema)
		require.NoError(t, err)
		require.Len(t, attrs, 2)

		var sshHostAttr Attribute
		for _, a := range attrs {
			if a.Name == "ssh_host" {
				sshHostAttr = a
			}
		}
		require.Len(t, sshHostAttr.Conditions, 1)
		assert.Equal(t, "ssh", sshHostAttr.Conditions[0].Field)
		assert.Equal(t, true, sshHostAttr.Conditions[0].Value)
		assert.Contains(t, sshHostAttr.Description, "Only applicable when")
	})

	t.Run("merges duplicate attributes from multiple conditions", func(t *testing.T) {
		// A field that appears under multiple enum values.
		props := jsonschema.NewProperties()
		props.Set("mode", &jsonschema.Schema{
			Type: "string",
			Enum: []interface{}{"a", "b"},
		})

		branch1Props := jsonschema.NewProperties()
		branch1Props.Set("mode", &jsonschema.Schema{Enum: []interface{}{"a"}})
		branch1Props.Set("shared_field", &jsonschema.Schema{Type: "string", Title: "Shared"})

		branch2Props := jsonschema.NewProperties()
		branch2Props.Set("mode", &jsonschema.Schema{Enum: []interface{}{"b"}})
		branch2Props.Set("shared_field", &jsonschema.Schema{Type: "string", Title: "Shared"})

		schema := &jsonschema.Schema{
			Type:       "object",
			Properties: props,
			DependentSchemas: map[string]*jsonschema.Schema{
				"mode": {
					OneOf: []*jsonschema.Schema{
						{Properties: branch1Props},
						{Properties: branch2Props},
					},
				},
			},
		}

		attrs, err := attributesForJSONSchema(schema)
		require.NoError(t, err)
		require.Len(t, attrs, 2, "shared_field should be merged, not duplicated")

		var shared Attribute
		for _, a := range attrs {
			if a.Name == "shared_field" {
				shared = a
			}
		}
		require.Len(t, shared.Conditions, 2)
		assert.Contains(t, shared.Description, `one of "a", "b"`)
	})
}

func TestTfAttr(t *testing.T) {
	t.Run("handles strings", func(t *testing.T) {
		attr, err := tfAttr("key", &jsonschema.Schema{Type: "string"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "key", attr.AttrName)
		assert.Equal(t, "schema.StringAttribute", attr.AttrType)
		assert.Equal(t, "types.StringType", attr.AttrReadType)
	})

	t.Run("handles arrays of primitives", func(t *testing.T) {
		attr, err := tfAttr("key", &jsonschema.Schema{
			Type:  "array",
			Items: &jsonschema.Schema{Type: "string"},
		}, nil)
		require.NoError(t, err)
		assert.Equal(t, "key", attr.AttrName)
		assert.Equal(t, "schema.SetAttribute", attr.AttrType)
		assert.Equal(t, "types.SetType", attr.AttrReadType)
		assert.Equal(t, "types.SetType", attr.AttrReadType)
		assert.Equal(t, "types.StringType", pointer.Get(attr.Elem).AttrReadType)
	})

	t.Run("handles arrays of objects", func(t *testing.T) {
		props := jsonschema.NewProperties()
		props.Set("name", &jsonschema.Schema{Type: "string"})
		props.Set("age", &jsonschema.Schema{Type: "integer"})

		attr, err := tfAttr("key", &jsonschema.Schema{
			Type: "array",
			Items: &jsonschema.Schema{
				Type:       "object",
				Properties: props,
			},
		}, nil)
		require.NoError(t, err)
		assert.Equal(t, "key", attr.AttrName)
		assert.Equal(t, "schema.SetNestedAttribute", attr.AttrType)
		assert.Equal(t, "types.SetType", attr.AttrReadType)
		assert.Equal(t, "types.ObjectType", pointer.Get(attr.Elem).AttrReadType)

		require.Len(t, attr.Elem.Attributes, 2)
	})

	t.Run("handles array properites of objects", func(t *testing.T) {
		elemProps := jsonschema.NewProperties()
		elemProps.Set("name", &jsonschema.Schema{Type: "string"})
		elemProps.Set("age", &jsonschema.Schema{Type: "integer"})

		props := jsonschema.NewProperties()
		props.Set("an_array", &jsonschema.Schema{
			Type: "array",
			Items: &jsonschema.Schema{
				Type:       "object",
				Properties: elemProps,
			},
		})

		attr, err := tfAttr("key", &jsonschema.Schema{
			Type:       "object",
			Properties: props,
		}, nil)

		require.NoError(t, err)
		assert.Equal(t, "key", attr.AttrName)
		assert.Equal(t, "schema.SingleNestedAttribute", attr.AttrType)
		assert.Equal(t, "types.ObjectType", attr.AttrReadType)
		require.Len(t, attr.Attributes, 1)

		arrayAttr := attr.Attributes[0]
		assert.Equal(t, "an_array", arrayAttr.AttrName)
		assert.Equal(t, "schema.SetNestedAttribute", arrayAttr.AttrType)
		assert.Equal(t, "types.SetType", arrayAttr.AttrReadType)
		assert.Equal(t, "types.ObjectType", pointer.Get(arrayAttr.Elem).AttrReadType)

		require.Len(t, arrayAttr.Elem.Attributes, 2)
		nameAttr := arrayAttr.Elem.Attributes[0]
		assert.Equal(t, "name", nameAttr.AttrName)
		assert.Equal(t, "schema.StringAttribute", nameAttr.AttrType)
	})
}
