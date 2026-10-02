package provider

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/require"
)

// Reuse an existing connection without placing it under Terraform ownership.
// With no override, retain the standalone Postgres fixture used in CI.
func syncTestPostgresConfig(t *testing.T) postgresTestConfig {
	t.Helper()
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("acceptance tests require TF_ACC=1")
	}
	testAccPreCheck(t)
	if os.Getenv("POLYTOMIC_SYNC_TEST_CONNECTION_ID") != "" {
		return postgresTestConfig{}
	}
	return testPostgresConfig(t)
}

func syncTestConfig(t *testing.T, source string) string {
	t.Helper()
	id := os.Getenv("POLYTOMIC_SYNC_TEST_CONNECTION_ID")
	if id == "" {
		return source
	}
	require.True(t, APIKey(), "existing connection override requires an organization API key")
	file, diags := hclwrite.ParseConfig([]byte(source), "sync-test.tf", hcl.InitialPos)
	require.False(t, diags.HasErrors(), "%s", diags.Error())
	for _, block := range file.Body().Blocks() {
		labels := block.Labels()
		if block.Type() == "resource" && len(labels) == 2 && labels[0] == "polytomic_postgresql_connection" && labels[1] == "test" {
			file.Body().RemoveBlock(block)
		}
	}
	return strings.ReplaceAll(string(file.Bytes()), "polytomic_postgresql_connection.test.id", strconv.Quote(id))
}
