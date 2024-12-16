package client

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/polytomic/polytomic-go"
	ptclient "github.com/polytomic/polytomic-go/client"
	ptoption "github.com/polytomic/polytomic-go/option"
)

const (
	UserAgent = "polytomic-terraform-provider"
)

// Provider is used to construct Polytomic clients based on the configured
// authentication mechanism.
type Provider struct {
	DeploymentKey string
	DeploymentURL string

	PartnerKey       string
	OrganizationUser string

	APIKey string

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

func WithPartnerKey(key, orgUser string) ProviderOpt {
	return func(p *Provider) {
		p.PartnerKey = key
		p.OrganizationUser = orgUser
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

func client(deploymentURL string, opts ...ptoption.RequestOption) *ptclient.Client {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 6

	return ptclient.NewClient(
		append([]ptoption.RequestOption{
			ptoption.WithBaseURL(deploymentURL),
			ptoption.WithHTTPHeader(http.Header{
				"User-Agent": []string{UserAgent},
			}),
			ptoption.WithHTTPClient(rc.StandardClient()),
		}, opts...)...,
	)
}

func (p *Provider) Client(org string) (*ptclient.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	orgID := uuid.Nil
	if org != "" {
		var err error
		orgID, err = uuid.Parse(org)
		if err != nil {
		}
	}
	if client, ok := p.clients[orgID]; ok {
		return client, nil
	}

	if p.APIKey != "" {
		// construct a client with the API key
		return client(p.DeploymentURL, ptoption.WithToken(p.APIKey)), nil
	}
	if p.PartnerKey != "" && orgID != uuid.Nil {
		ctx := context.Background()
		// ensure that the org user exists in the organization and return a client using its key
		pc, err := p.PartnerClient()
		if err != nil {
			panic(err)
		}

		userID := uuid.Nil
		if uid, err := uuid.Parse(p.OrganizationUser); err == nil {
			userID = uid
		} else {
			users, err := pc.Users.List(ctx, org)
			if err != nil {
				return nil, err
			}
			for _, user := range users.Data {
				if strings.ToLower(pointer.Get(user.Email)) == strings.ToLower(p.OrganizationUser) {
					userID = uuid.MustParse(*user.Id)
					break
				}
			}
		}
		if userID == uuid.Nil {
		}

		key, err := pc.Users.CreateApiKey(ctx, org, userID.String(), &polytomic.UsersCreateApiKeyRequest{Force: pointer.To(true)})
		if err != nil {
			return nil, err
		}
		p.clients[orgID] = client(p.DeploymentURL, ptoption.WithToken(key.Data.String()))
		return p.clients[orgID], nil
	}
	if p.DeploymentKey != "" {
		// construct a client with the deployment key
		return client(p.DeploymentURL,
			ptoption.WithHTTPHeader(http.Header{
				"Authorization": []string{"Basic " + basicAuth(p.DeploymentKey, "")},
			}),
		), nil
	}

	return nil, errors.New("No valid authentication method found")
}

func (p *Provider) PartnerClient() (*ptclient.Client, error) {
	if p.PartnerKey == "" && p.DeploymentKey == "" {
		return nil, errors.New("Partner Key is required")
	}
	if p.PartnerKey != "" {
		return client(p.DeploymentURL, ptoption.WithToken(p.PartnerKey)), nil
	}
	return client(p.DeploymentURL,
		ptoption.WithHTTPHeader(http.Header{
			"Authorization": []string{"Basic " + basicAuth(p.DeploymentKey, "")},
		}),
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
