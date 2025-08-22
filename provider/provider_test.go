package provider

import (
	"os"
	"testing"

	ptclient "github.com/polytomic/polytomic-go/client"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/providerclient"
	"github.com/stretchr/testify/require"
)

func testAccPreCheck(t *testing.T) {
	TestAccPreCheck(t)
}

func testClient(t *testing.T, org string) *ptclient.Client {
	t.Helper()

	provider, err := providerclient.NewClientProvider(
		providerclient.WithDeploymentKey(os.Getenv(PolytomicDeploymentKey)),
		providerclient.WithDeploymentURL(os.Getenv(PolytomicDeploymentURL)),
		providerclient.WithAPIKey(os.Getenv(PolytomicAPIKey)),
	)
	require.NoError(t, err)
	c, err := provider.Client(org)
	require.NoError(t, err)
	return c
}

func testPartnerClient(t *testing.T) *ptclient.Client {
	t.Helper()

	provider, err := providerclient.NewClientProvider(
		providerclient.WithDeploymentKey(os.Getenv(PolytomicDeploymentKey)),
		providerclient.WithDeploymentURL(os.Getenv(PolytomicDeploymentURL)),
		providerclient.WithAPIKey(os.Getenv(PolytomicAPIKey)),
	)
	require.NoError(t, err)
	c, err := provider.PartnerClient()
	require.NoError(t, err)
	return c
}
