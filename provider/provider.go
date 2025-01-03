package provider

import (
	"cmp"
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/providervalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/connections"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/providerclient"
)

const (
	// Name is the name of the provider.
	Name = "polytomic"

	//PolytomicDeploymentKey is the environment variable name for the Polytomic deployment key
	PolytomicDeploymentKey = "POLYTOMIC_DEPLOYMENT_KEY"
	//PolytomicAPIKey is the environment variable name for the Polytomic API key
	PolytomicAPIKey = "POLYTOMIC_API_KEY"
	//PolytomicDeploymentURL is the environment variable name for the Polytomic deployment URL
	PolytomicDeploymentURL = "POLYTOMIC_DEPLOYMENT_URL"

	PolytomicDefaultURL = "app.polytomic.com"
)

var (
	_ provider.ProviderWithConfigValidators = (*Provider)(nil)
)

// ProviderData holds the provider configuration, which is used to construct
// Polytomic clients.
type ProviderData struct {
	DeploymentKey types.String `tfsdk:"deployment_api_key"`
	DeploymentUrl types.String `tfsdk:"deployment_url"`

	PartnerKey       types.String `tfsdk:"partner_key"`
	OrganizationUser types.String `tfsdk:"organization_user"`

	APIKey types.String `tfsdk:"api_key"`
}

// Provider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type Provider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = Name
	resp.Version = p.version
}

func (p *Provider) ConfigValidators(context.Context) []provider.ConfigValidator {
	return []provider.ConfigValidator{
		providervalidator.Conflicting(
			path.MatchRoot("deployment_api_key"),
			path.MatchRoot("api_key"),
			path.MatchRoot("partner_key"),
		),
		providervalidator.RequiredTogether(
			path.MatchRoot("deployment_api_key"),
			path.MatchRoot("deployment_url"),
		),
		providervalidator.RequiredTogether(
			path.MatchRoot("partner_key"),
			path.MatchRoot("organization_user"),
		),
	}
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProviderData
	resp.Diagnostics.Append(
		req.Config.Get(ctx, &data)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment := cmp.Or(
		data.DeploymentUrl.ValueString(),
		os.Getenv(PolytomicDeploymentURL),
		PolytomicDefaultURL,
	)
	if !strings.HasPrefix(strings.ToLower(deployment), "http") {
		deployment = "https://" + deployment
	}
	deploymentURL, err := url.Parse(deployment)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing deployment URL", err.Error())
		return
	}
	if deploymentURL.Scheme == "" {
		deploymentURL.Scheme = "https"
	}
	providerOpts := []providerclient.ProviderOpt{
		providerclient.WithDeploymentURL(deploymentURL.String()),
		providerclient.WithDeploymentKey(
			cmp.Or(
				data.DeploymentKey.ValueString(),
				os.Getenv(PolytomicDeploymentKey),
			),
		),
		providerclient.WithPartnerKey(data.PartnerKey.ValueString(), data.OrganizationUser.ValueString()),
		providerclient.WithAPIKey(
			cmp.Or(
				data.APIKey.ValueString(),
				os.Getenv(PolytomicAPIKey),
			),
		),
	}

	clientProvider, err := providerclient.NewClientProvider(providerOpts...)
	if err != nil {
		resp.Diagnostics.AddError("Error configuring provider", err.Error())
		return
	}

	resp.DataSourceData = clientProvider
	resp.ResourceData = clientProvider
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	resourceList := []func() resource.Resource{
		func() resource.Resource { return &organizationResource{} },
		func() resource.Resource { return &userResource{} },
		func() resource.Resource { return &roleResource{} },
		func() resource.Resource { return &policyResource{} },
		func() resource.Resource { return &modelResource{} },
		func() resource.Resource { return &bulkSyncResource{} },
		func() resource.Resource { return &syncResource{} },
	}
	all := append(connections.Resources, resourceList...)
	return all

}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	datasources := []func() datasource.DataSource{
		func() datasource.DataSource { return &bulkSourceDatasource{} },
		func() datasource.DataSource { return &bulkDestinationDatasource{} },
		func() datasource.DataSource { return &identityDatasource{} },
	}
	all := append(connections.Datasources, datasources...)
	return all
}

func (p *Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"deployment_url": schema.StringAttribute{
				MarkdownDescription: "Polytomic deployment URL (defaults to app.polytomic.com)",
				Optional:            true,
			},
			"deployment_api_key": schema.StringAttribute{
				MarkdownDescription: "Polytomic deployment key",
				Optional:            true,
				Sensitive:           true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Polytomic API key",
				Optional:            true,
				Sensitive:           true,
			},
			"partner_key": schema.StringAttribute{
				MarkdownDescription: "Polytomic partner key",
				Optional:            true,
				Sensitive:           true,
			},
			"organization_user": schema.StringAttribute{
				MarkdownDescription: "Polytomic organization user; required if `partner_key` is set.",
				Optional:            true,
				Sensitive:           false,
			},
		},
	}

}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{
			version: version,
		}
	}
}
