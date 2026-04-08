package provider

import (
	"database/sql"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	_ "github.com/lib/pq"
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

func testPostgresConfigFromEnv(t *testing.T) (postgresTestConfig, bool) {
	t.Helper()

	if os.Getenv("POLYTOMIC_TEST_PG_USERNAME") == "" {
		return postgresTestConfig{}, false
	}

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

	if cfg.Password == "" {
		t.Fatalf("POLYTOMIC_TEST_PG_PASSWORD must be set when POLYTOMIC_TEST_PG_USERNAME is set")
	}

	// Ensure fixture tables exist for BYOP Postgres.
	ensureFixtureTables(t, cfg)

	return cfg, true
}

func getenvOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}

// fixtureSQL creates the schemas and tables expected by acceptance tests.
const fixtureSQL = `
CREATE SCHEMA IF NOT EXISTS polytomic;

CREATE TABLE IF NOT EXISTS polytomic.sync_test_source (
    email TEXT PRIMARY KEY,
    name TEXT
);

INSERT INTO polytomic.sync_test_source (email, name)
VALUES ('test@example.com', 'Test User')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS polytomic.sync_test_target (
    email TEXT PRIMARY KEY,
    name TEXT
);
`

// ensureFixtureTables connects to the given Postgres instance and runs
// fixtureSQL to create the schemas and seed data needed by acceptance tests.
func ensureFixtureTables(t *testing.T, cfg postgresTestConfig) {
	t.Helper()

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("connecting to postgres for fixture setup: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(fixtureSQL); err != nil {
		t.Fatalf("running fixture SQL: %v", err)
	}
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
