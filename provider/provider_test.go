package provider

import (
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	ptclient "github.com/polytomic/polytomic-go/client"
	ptoption "github.com/polytomic/polytomic-go/option"
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

func testClient() *ptclient.Client {
	deployURL := os.Getenv(PolytomicDeploymentURL)
	deployKey := os.Getenv(PolytomicDeploymentKey)
	apiKey := os.Getenv(PolytomicAPIKey)

	var client *ptclient.Client
	if deployKey != "" {
		client = ptclient.NewClient(
			ptoption.WithBaseURL(deployURL),
			ptoption.WithHTTPHeader(http.Header{
				"Authorization ": []string{"Basic " + basicAuth(deployKey, "")},
			}),
		)
	} else {
		client = ptclient.NewClient(
			ptoption.WithBaseURL(deployURL),
			ptoption.WithToken(apiKey),
		)
	}

	return client
}
