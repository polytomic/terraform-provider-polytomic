package provider

import (
	"html/template"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
	"github.com/stretchr/testify/require"
)

// TestAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	Name: providerserver.NewProtocol6WithError(New("test")()),
}

// TestAccPreCheck performs pre-checks for acceptance testing
func TestAccPreCheck(t *testing.T) {
	if os.Getenv(providerclient.PolytomicAPIKey) == "" && os.Getenv(providerclient.PolytomicDeploymentKey) == "" {
		t.Fatalf("%s or %s must be set for acceptance testing", providerclient.PolytomicAPIKey, providerclient.PolytomicDeploymentKey)
	}

	if os.Getenv(providerclient.PolytomicDeploymentURL) == "" {
		t.Fatalf("%s must be set for acceptance testing", providerclient.PolytomicDeploymentURL)
	}
}

// GetTestAccProtoV6ProviderFactories returns the provider factories for testing
func GetTestAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return TestAccProtoV6ProviderFactories
}

// APIKey returns true if the test is being run using an API key, rather than a
// deployment key.
func APIKey() bool {
	return os.Getenv(providerclient.PolytomicAPIKey) != "" && os.Getenv(providerclient.PolytomicDeploymentKey) == ""
}

type postgresTestConfig struct {
	Host     string
	Database string
	Username string
	Password string
	Port     int
}

func testPostgresConfig(t *testing.T) postgresTestConfig {
	t.Helper()

	cfg := postgresTestConfig{
		Host:     getenvOr("POLYTOMIC_TEST_PG_HOST", "postgres"),
		Database: getenvOr("POLYTOMIC_TEST_PG_DATABASE", "polytomic"),
		Username: os.Getenv("POLYTOMIC_TEST_PG_USERNAME"),
		Password: os.Getenv("POLYTOMIC_TEST_PG_PASSWORD"),
	}

	port, err := strconv.Atoi(getenvOr("POLYTOMIC_TEST_PG_PORT", "5432"))
	if err != nil {
		t.Fatalf("POLYTOMIC_TEST_PG_PORT must be a valid integer: %v", err)
	}
	cfg.Port = port

	if cfg.Username == "" {
		t.Fatalf("POLYTOMIC_TEST_PG_USERNAME must be set for PostgreSQL-backed acceptance tests")
	}

	if cfg.Password == "" {
		t.Fatalf("POLYTOMIC_TEST_PG_PASSWORD must be set for PostgreSQL-backed acceptance tests")
	}

	return cfg
}

func getenvOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

type TestCaseTfArgs struct {
	Name     string
	APIKey   bool
	Postgres postgresTestConfig
}

// TestCaseTfResource generates the Terraform configuration for a test case from
// the provided template.
func TestCaseTfResource(t *testing.T, tmplText string, args TestCaseTfArgs) string {
	t.Helper()

	tmpl := template.Must(template.New("testcase").Parse(tmplText))

	var result strings.Builder
	if err := tmpl.Execute(&result, args); err != nil {
		require.NoError(t, err, "error executing template")
	}

	return result.String()

}
