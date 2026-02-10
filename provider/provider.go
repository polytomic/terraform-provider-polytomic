package provider

import (
	"cmp"
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/providervalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/connections"
)

const (
	// Name is the name of the provider.
	Name = "polytomic"

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

	PartnerKey types.String `tfsdk:"partner_key"`

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

	clientProvider, err := providerclient.NewClientProvider(
		providerclient.Options{
			DeploymentURL: cmp.Or(
				data.DeploymentUrl.ValueString(),
				os.Getenv(providerclient.PolytomicDeploymentURL),
				PolytomicDefaultURL,
			),
			DeploymentKey: cmp.Or(
				data.DeploymentKey.ValueString(),
				os.Getenv(providerclient.PolytomicDeploymentKey),
			),
			PartnerKey: cmp.Or(
				data.PartnerKey.ValueString(),
				os.Getenv(providerclient.PolytomicPartnerKey),
			),
			APIKey: cmp.Or(
				data.APIKey.ValueString(),
				os.Getenv(providerclient.PolytomicAPIKey),
			),
		},
	)
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
		func() resource.Resource { return &globalErrorSubscribersResource{} },
		func() resource.Resource { return &modelResource{} },
		func() resource.Resource { return &bulkSyncResource{} },
		func() resource.Resource { return &syncResource{} },
		NewConnectionSchemaPrimaryKeysResource,
	}
	all := append(connections.Resources, resourceList...)
	return all
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	datasources := []func() datasource.DataSource{
		func() datasource.DataSource { return &bulkSourceDatasource{} },
		func() datasource.DataSource { return &bulkDestinationDatasource{} },
		func() datasource.DataSource { return &identityDatasource{} },
		NewConnectionSchemaDataSource,
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
