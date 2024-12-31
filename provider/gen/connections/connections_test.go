package connections

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
