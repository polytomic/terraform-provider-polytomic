package providerclient

import (
	"cmp"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/polytomic/polytomic-go"
	ptclient "github.com/polytomic/polytomic-go/client"
	ptoption "github.com/polytomic/polytomic-go/option"
)

const (
	UserAgent    = "polytomic-terraform-provider"
	ErrorSummary = "Client Error"

	//PolytomicDeploymentKey is the environment variable name for the Polytomic deployment key
	PolytomicDeploymentKey = "POLYTOMIC_DEPLOYMENT_KEY"
	//PolytomicAPIKey is the environment variable name for the Polytomic API key
	PolytomicAPIKey = "POLYTOMIC_API_KEY"
	//PolytomicPartnerKey is the environment variable name for the Polytomic partner key
	PolytomicPartnerKey = "POLYTOMIC_PARTNER_KEY"
	//PolytomicDeploymentURL is the environment variable name for the Polytomic deployment URL
	PolytomicDeploymentURL = "POLYTOMIC_DEPLOYMENT_URL"
)

type Options struct {
	DeploymentKey string
	DeploymentURL string
	PartnerKey    string
	APIKey        string
}

func (o Options) Validate() error {
	if o.DeploymentKey == "" && o.DeploymentURL == "" && o.PartnerKey == "" && o.APIKey == "" {
		return fmt.Errorf("API Key, Deployment Key, or Partner Key must be set")
	}
	return nil
}

// OptionsFromEnv returns the provider client Options configured using the
// environment.
func OptionsFromEnv() Options {
	return Options{
		DeploymentKey: os.Getenv(PolytomicDeploymentKey),
		DeploymentURL: os.Getenv(PolytomicDeploymentURL),
		PartnerKey:    os.Getenv(PolytomicPartnerKey),
		APIKey:        os.Getenv(PolytomicAPIKey),
	}
}

// Provider is used to construct Polytomic clients based on the configured
// authentication mechanism.
type Provider struct {
	opts Options

	mu      sync.Mutex
	clients map[uuid.UUID]*ptclient.Client
}

func NewClientProvider(opts Options) (*Provider, error) {
	// Normalize deployment URL
	if opts.DeploymentURL == "" {
		opts.DeploymentURL = "app.polytomic.com"
	}

	// Add https:// scheme if no scheme is present
	if !strings.HasPrefix(strings.ToLower(opts.DeploymentURL), "http") {
		opts.DeploymentURL = "https://" + opts.DeploymentURL
	}

	// Parse and validate the URL
	deploymentURL, err := url.Parse(opts.DeploymentURL)
	if err != nil {
		return nil, fmt.Errorf("invalid deployment URL: %w", err)
	}

	// Ensure scheme is set (redundant but defensive)
	if deploymentURL.Scheme == "" {
		deploymentURL.Scheme = "https"
	}

	// Remove all trailing slashes from the full URL
	opts.DeploymentURL = strings.TrimRight(deploymentURL.String(), "/")

	p := &Provider{
		opts:    opts,
		clients: map[uuid.UUID]*ptclient.Client{},
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Validate checks the provider configuration for required fields.
func (p *Provider) Validate() error {
	return p.opts.Validate()
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

	if p.opts.APIKey != "" {
		if orgID != uuid.Nil {
			// confirm that the requested org ID matches the org ID for this key
			orgs, err := p.ListOrganizations(context.Background())
			if err != nil {
				return nil, fmt.Errorf("failed to list organizations for API key: %w", err)
			}
			if len(orgs) < 1 || pointer.Get(orgs[0].Id) != orgID.String() {
				return nil, fmt.Errorf("API key does not have access to organization %s", orgID)
			}
		}

		return p.client()
	}

	if orgID == uuid.Nil {
		return nil, errors.New("organization ID must be specified with partner or deployment key")
	}
	if p.opts.PartnerKey != "" || p.opts.DeploymentKey != "" {
		if orgID == uuid.Nil {
			return p.PartnerClient()
		}

		headers := http.Header{
			"User-Agent": []string{UserAgent},
			// Use Basic Auth with organization ID as username and partner key as password
			"Authorization": []string{
				"Basic " + basicAuth(orgID.String(), cmp.Or(p.opts.PartnerKey, p.opts.DeploymentKey)),
			},
		}

		p.clients[orgID] = ptclient.NewClient(
			ptoption.WithBaseURL(p.opts.DeploymentURL),
			ptoption.WithHTTPHeader(headers),
		)
		return p.clients[orgID], nil
	}

	return nil, errors.New("no valid authentication method found")
}

// client returns a Polytomic HTTP client configured using the provider options.
func (p *Provider) client() (*ptclient.Client, error) {
	headers := http.Header{
		"User-Agent": []string{UserAgent},
	}

	if p.opts.APIKey != "" {
		// construct a client with the API key
		return ptclient.NewClient(
			ptoption.WithBaseURL(p.opts.DeploymentURL),
			ptoption.WithToken(p.opts.APIKey),
			ptoption.WithHTTPHeader(headers),
		), nil
	}

	if p.opts.PartnerKey != "" || p.opts.DeploymentKey != "" {
		// For partner client without specific org, use partner key as bearer token
		return ptclient.NewClient(
			ptoption.WithBaseURL(p.opts.DeploymentURL),
			ptoption.WithToken(
				cmp.Or(p.opts.PartnerKey, p.opts.DeploymentKey),
			),
			ptoption.WithHTTPHeader(headers),
		), nil
	}

	return nil, errors.New("no valid authentication method found")
}

func (p *Provider) PartnerClient() (*ptclient.Client, error) {
	if p.opts.PartnerKey == "" && p.opts.DeploymentKey == "" {
		return nil, errors.New("partner key is required")
	}
	return p.client()
}

// ListOrganizations returns all organizations accessible via the configured
// authentication method.
func (ic *Provider) ListOrganizations(ctx context.Context) ([]*polytomic.Organization, error) {
	c, err := ic.client()
	if err != nil {
		return nil, fmt.Errorf("failed to get Polytomic client: %w", err)
	}

	orgsResp, err := c.Organization.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	return orgsResp.Data, nil
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
	auth := fmt.Sprintf("%s:%s",
		strings.TrimSpace(username),
		strings.TrimSpace(password),
	)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
