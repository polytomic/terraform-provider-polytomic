package provider

import (
	"testing"

	ptclient "github.com/polytomic/polytomic-go/client"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
	"github.com/stretchr/testify/require"
)

func testAccPreCheck(t *testing.T) {
	TestAccPreCheck(t)
}

func testClient(t *testing.T, org string) *ptclient.Client {
	t.Helper()

	provider, err := providerclient.NewClientProvider(
		providerclient.OptionsFromEnv(),
	)
	require.NoError(t, err)
	c, err := provider.Client(org)
	require.NoError(t, err)
	return c
}

func testPartnerClient(t *testing.T) *ptclient.Client {
	t.Helper()

	provider, err := providerclient.NewClientProvider(
		providerclient.OptionsFromEnv(),
	)
	require.NoError(t, err)
	c, err := provider.PartnerClient()
	require.NoError(t, err)
	return c
}
