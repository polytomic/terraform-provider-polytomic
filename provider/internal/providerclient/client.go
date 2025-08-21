package providerclient

import (
	"cmp"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	ptclient "github.com/polytomic/polytomic-go/client"
	ptoption "github.com/polytomic/polytomic-go/option"
)

const (
	UserAgent    = "polytomic-terraform-provider"
	ErrorSummary = "Client Error"
)

// Provider is used to construct Polytomic clients based on the configured
// authentication mechanism.
type Provider struct {
	DeploymentKey string
	DeploymentURL string
	PartnerKey    string
	APIKey        string

	mu      sync.Mutex
	clients map[uuid.UUID]*ptclient.Client
}

type ProviderOpt func(*Provider)

func WithDeploymentKey(key string) ProviderOpt {
	return func(p *Provider) {
		p.DeploymentKey = key
	}
}

func WithDeploymentURL(url string) ProviderOpt {
	return func(p *Provider) {
		p.DeploymentURL = url
	}
}

func WithPartnerKey(key string) ProviderOpt {
	return func(p *Provider) {
		p.PartnerKey = key
	}
}

func WithAPIKey(key string) ProviderOpt {
	return func(p *Provider) {
		p.APIKey = key
	}
}

func NewClientProvider(opts ...ProviderOpt) (*Provider, error) {
	p := &Provider{
		clients: map[uuid.UUID]*ptclient.Client{},
	}
	for _, opt := range opts {
		opt(p)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Validate checks the provider configuration for required fields.
func (p *Provider) Validate() error {
	if p.APIKey == "" && p.DeploymentKey == "" && p.PartnerKey == "" {
		return fmt.Errorf("API Key, Deployment Key, or Partner Key must be set")
	}
	return nil
}

func (p *Provider) Client(org string) (*ptclient.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	orgID := uuid.Nil
	if org != "" {
		var err error
		orgID, err = uuid.Parse(org)
		if err != nil {
			return nil, fmt.Errorf("invalid organization ID %s: %w", org, err)
		}
	}
	if client, ok := p.clients[orgID]; ok {
		return client, nil
	}

	headers := http.Header{
		"User-Agent": []string{UserAgent},
	}
	if p.APIKey != "" {
		// construct a client with the API key
		return ptclient.NewClient(
			ptoption.WithBaseURL(p.DeploymentURL),
			ptoption.WithToken(p.APIKey),
			ptoption.WithHTTPHeader(headers),
		), nil
	}
	if orgID == uuid.Nil {
		return nil, errors.New("organization ID must be specified with partner or deployment key")
	}
	if p.PartnerKey != "" || p.DeploymentKey != "" {
		// Use Basic Auth with organization ID as username and partner key as password
		headers["Authorization"] = []string{
			"Basic " + basicAuth(orgID.String(), cmp.Or(p.PartnerKey, p.DeploymentKey)),
		}

		p.clients[orgID] = ptclient.NewClient(
			ptoption.WithBaseURL(p.DeploymentURL),
			ptoption.WithHTTPHeader(headers),
		)
		return p.clients[orgID], nil
	}

	return nil, errors.New("no valid authentication method found")
}

func (p *Provider) PartnerClient() (*ptclient.Client, error) {
	if p.PartnerKey == "" && p.DeploymentKey == "" {
		return nil, errors.New("partner key is required")
	}
	headers := http.Header{
		"User-Agent": []string{UserAgent},
	}
	// For partner client without specific org, use partner key as bearer token
	return ptclient.NewClient(
		ptoption.WithBaseURL(p.DeploymentURL),
		ptoption.WithToken(
			cmp.Or(p.PartnerKey, p.DeploymentKey),
		),
		ptoption.WithHTTPHeader(headers),
	), nil
}

// GetClient returns a Polytomic client from the provider data for the specified
// organization.
func GetProvider(data any, diags diag.Diagnostics) *Provider {
	// Prevent panic if the provider has not been configured.
	if data == nil {
		return nil
	}

	clients, ok := data.(*Provider)
	if !ok {
		diags.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *polytomic.Client, got: %T. Please report this issue to the provider developers.", data),
		)
		return nil
	}

	return clients
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
