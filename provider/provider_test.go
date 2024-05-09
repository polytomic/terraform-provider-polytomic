package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/polytomic/polytomic-go"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	Name: providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.

	if os.Getenv(PolytomicAPIKey) == "" && os.Getenv(PolytomicDeploymentKey) == "" {
		t.Fatalf("%s or %s must be set for acceptance testing", PolytomicAPIKey, PolytomicDeploymentKey)
	}

	if os.Getenv(PolytomicDeploymentURL) == "" {
		t.Fatalf("%s must be set for acceptance testing", PolytomicDeploymentURL)
	}
}

func testClient() *polytomic.Client {
	deployURL := os.Getenv(PolytomicDeploymentURL)
	deployKey := os.Getenv(PolytomicDeploymentKey)
	apiKey := os.Getenv(PolytomicAPIKey)

	var client *polytomic.Client
	if deployKey != "" {
		client = polytomic.NewClient(
			deployURL,
			polytomic.DeploymentKey(deployKey),
		)
	} else {
		client = polytomic.NewClient(
			deployURL,
			polytomic.APIKey(apiKey),
		)
	}

	return client
}
