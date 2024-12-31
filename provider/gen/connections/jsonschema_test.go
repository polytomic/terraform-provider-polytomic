package connections

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalJSONSchema(t *testing.T) {
	csvBytes, err := os.ReadFile("./connectiontypes/csv.json")
	require.NoError(t, err)
	csvJSON := map[string]interface{}{}
	err = json.Unmarshal(csvBytes, &csvJSON)
	require.NoError(t, err)

	schema, err := unmarshalJSONSchema(csvJSON)
	require.NoError(t, err)

	assert.Equal(t, "object", schema.Type)
	assert.Equal(t, 4, schema.Properties.Len())

	authProp, ok := schema.Properties.Get("auth")
	require.True(t, ok)
	assert.Equal(t, "object", authProp.Type)

	oauthProp, ok := authProp.Properties.Get("oauth")
	require.True(t, ok)
	assert.Equal(t, "object", oauthProp.Type)

	extraData, ok := oauthProp.Properties.Get("extra_form_data")
	require.True(t, ok)
	assert.Equal(t, "array", extraData.Type)

	extraDataItems := extraData.Items
	require.NotNil(t, extraDataItems)
	assert.Equal(t, "object", extraDataItems.Type)

	name, ok := extraDataItems.Properties.Get("name")
	require.True(t, ok)
	assert.Equal(t, "string", name.Type)
}
