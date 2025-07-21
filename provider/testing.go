package provider

import (
	"html/template"
	"os"
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

type TestCaseTfArgs struct {
	Name   string
	APIKey bool
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
