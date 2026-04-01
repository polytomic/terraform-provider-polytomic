package connections

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalJSONSchemaWithDependentSchemas(t *testing.T) {
	input := `{
		"type": "object",
		"properties": {
			"auth_mode": {"type": "string", "enum": ["key", "role"]}
		},
		"dependentSchemas": {
			"auth_mode": {
				"oneOf": [
					{
						"properties": {
							"auth_mode": {"enum": ["key"]},
							"api_key": {"type": "string", "title": "API Key"}
						},
						"required": ["api_key"]
					},
					{
						"properties": {
							"auth_mode": {"enum": ["role"]},
							"role_arn": {"type": "string", "title": "Role ARN"}
						}
					}
				]
			}
		}
	}`
	var raw map[string]interface{}
	err := json.Unmarshal([]byte(input), &raw)
	require.NoError(t, err)

	schema, err := unmarshalJSONSchema(raw)
	require.NoError(t, err)

	assert.Equal(t, "object", schema.Type)
	require.Len(t, schema.DependentSchemas, 1)

	depSchema, ok := schema.DependentSchemas["auth_mode"]
	require.True(t, ok)
	require.Len(t, depSchema.OneOf, 2)

	// First oneOf branch should have api_key.
	branch1 := depSchema.OneOf[0]
	apiKey, ok := branch1.Properties.Get("api_key")
	require.True(t, ok)
	assert.Equal(t, "string", apiKey.Type)
	assert.Equal(t, "API Key", apiKey.Title)
	assert.Equal(t, []string{"api_key"}, branch1.Required)

	// Second oneOf branch should have role_arn.
	branch2 := depSchema.OneOf[1]
	roleArn, ok := branch2.Properties.Get("role_arn")
	require.True(t, ok)
	assert.Equal(t, "string", roleArn.Type)
}

func TestUnmarshalJSONSchemaWithIfThenElse(t *testing.T) {
	input := `{
		"type": "object",
		"properties": {
			"ssh": {"type": "boolean"}
		},
		"dependentSchemas": {
			"ssh": {
				"if": {
					"properties": {"ssh": {"const": true}}
				},
				"then": {
					"properties": {
						"ssh_host": {"type": "string", "title": "SSH Host"}
					},
					"required": ["ssh_host"]
				}
			}
		}
	}`
	var raw map[string]interface{}
	err := json.Unmarshal([]byte(input), &raw)
	require.NoError(t, err)

	schema, err := unmarshalJSONSchema(raw)
	require.NoError(t, err)

	depSchema := schema.DependentSchemas["ssh"]
	require.NotNil(t, depSchema.If)
	require.NotNil(t, depSchema.Then)

	// Verify the if-clause has const: true.
	ifProp, ok := depSchema.If.Properties.Get("ssh")
	require.True(t, ok)
	assert.Equal(t, true, ifProp.Const)

	// Verify the then-clause has ssh_host.
	sshHost, ok := depSchema.Then.Properties.Get("ssh_host")
	require.True(t, ok)
	assert.Equal(t, "string", sshHost.Type)
	assert.Equal(t, []string{"ssh_host"}, depSchema.Then.Required)
}

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
