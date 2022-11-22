package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
)

const (
	//PolytomicDeploymentKey is the environment variable name for the Polytomic deployment key
	PolytomicDeploymentKey = "POLYTOMIC_DEPLOYMENT_KEY"
	//PolytomicAPIKey is the environment variable name for the Polytomic API key
	PolytomicAPIKey = "POLYTOMIC_API_KEY"
	//PolytomicDeploymentURL is the environment variable name for the Polytomic deployment URL
	PolytomicDeploymentURL = "POLYTOMIC_DEPLOYMENT_URL"
)

const clientError = "Client Error"

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ provider.Provider             = &ptProvider{}
	_ provider.ProviderWithMetadata = &ptProvider{}
)

// ptProvider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type ptProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// providerData can be used to store data from the Terraform configuration.
type ProviderData struct {
	DeploymentKey types.String `tfsdk:"deployment_api_key"`
	DeploymentUrl types.String `tfsdk:"deployment_url"`
	APIKey        types.String `tfsdk:"api_key"`
}

func (p *ptProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "polytomic"
	resp.Version = p.version
}

func (p *ptProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProviderData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var deployURL, deployKey, apiKey string

	// If the deployment URL is not set in the provider configuration, check the environment
	if data.DeploymentKey.IsNull() {
		deployKey = os.Getenv(PolytomicDeploymentKey)
	} else {
		deployKey = data.DeploymentKey.ValueString()
	}

	if data.APIKey.IsNull() {
		apiKey = os.Getenv(PolytomicAPIKey)
	} else {
		apiKey = data.APIKey.ValueString()
	}

	if deployKey == "" && apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Polytomic Deployment or API Key",
			fmt.Sprintf(`Please set the "deployment_api_key" or "api_key" in the provider configuration or one of the %s or %s environment variables`, PolytomicDeploymentKey, PolytomicAPIKey),
		)
		return
	}

	// If the deployment URL is not set in the provider configuration, check the environment
	if data.DeploymentUrl.IsNull() {
		deployURL = os.Getenv(PolytomicDeploymentURL)
	} else {
		deployURL = data.DeploymentUrl.ValueString()
	}

	if deployURL == "" {
		resp.Diagnostics.AddError(
			"Missing Polytomic Deployment URL",
			fmt.Sprintf(`Please set the "deployment_url" in the provider configuration or the %s environment variable`, PolytomicDeploymentURL),
		)
		return
	}
	var client *polytomic.Client
	// Deployment key is the default and takes precedence
	if apiKey != "" && deployKey == "" {
		client = polytomic.NewClient(
			deployURL,
			polytomic.APIKey(apiKey),
		)
	} else {
		client = polytomic.NewClient(
			deployURL,
			polytomic.DeploymentKey(deployKey),
		)
	}

	resp.DataSourceData = client
	resp.ResourceData = client

}

func (p *ptProvider) Resources(ctx context.Context) []func() resource.Resource {
	resourceList := []func() resource.Resource{
		func() resource.Resource { return &organizationResource{} },
		func() resource.Resource { return &userResource{} },
		func() resource.Resource { return &modelResource{} },
		func() resource.Resource { return &bulkSyncResource{} },
		func() resource.Resource { return &bulkSyncSchemaResource{} },
	}
	all := append(connection_resources, resourceList...)
	return all

}

func (p *ptProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	datasources := []func() datasource.DataSource{
		func() datasource.DataSource { return &bulkSourceDatasource{} },
		func() datasource.DataSource { return &bulkDestinationDatasource{} },
	}
	all := append(connection_datasources, datasources...)
	return all
}

func (p *ptProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"deployment_url": {
				MarkdownDescription: "Polytomic deployment URL",
				Type:                types.StringType,
				Optional:            true,
			},
			"deployment_api_key": {
				MarkdownDescription: "Polytomic deployment key",
				Type:                types.StringType,
				Optional:            true,
				Sensitive:           true,
			},
			"api_key": {
				MarkdownDescription: "Polytomic API key",
				Type:                types.StringType,
				Optional:            true,
				Sensitive:           true,
			},
		},
	}, nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ptProvider{
			version: version,
		}
	}
}
