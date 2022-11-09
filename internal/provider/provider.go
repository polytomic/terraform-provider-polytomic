package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
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

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.Provider = &ptProvider{}

// ptProvider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type ptProvider struct {
	client *polytomic.Client

	// configured is set to true at the end of the Configure method.
	// This can be used in Resource and DataSource implementations to verify
	// that the provider was previously configured.
	configured bool

	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// providerData can be used to store data from the Terraform configuration.
type providerData struct {
	DeploymentKey types.String `tfsdk:"deployment_api_key"`
	DeploymentUrl types.String `tfsdk:"deployment_url"`
	APIKey        types.String `tfsdk:"api_key"`
}

func (p *ptProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var deployURL, deployKey, apiKey string

	// If the deployment URL is not set in the provider configuration, check the environment
	if data.DeploymentKey.Null {
		deployKey = os.Getenv(PolytomicDeploymentKey)
	} else {
		deployKey = data.DeploymentKey.Value
	}

	if data.APIKey.Null {
		apiKey = os.Getenv(PolytomicAPIKey)
	} else {
		apiKey = data.APIKey.Value
	}

	if deployKey == "" && apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Polytomic Deployment or API Key",
			fmt.Sprintf(`Please set the "deployment_api_key" or "api_key" in the provider configuration or one of the %s or %s environment variables`, PolytomicDeploymentKey, PolytomicAPIKey),
		)
		return
	}

	// If the deployment URL is not set in the provider configuration, check the environment
	if data.DeploymentUrl.Null {
		deployURL = os.Getenv(PolytomicDeploymentURL)
	} else {
		deployURL = data.DeploymentUrl.Value
	}

	if deployURL == "" {
		resp.Diagnostics.AddError(
			"Missing Polytomic Deployment URL",
			fmt.Sprintf(`Please set the "deployment_url" in the provider configuration or the %s environment variable`, PolytomicDeploymentURL),
		)
		return
	}
	// Deployment key is the default and takes precedence
	if apiKey != "" && deployKey == "" {
		p.client = polytomic.NewClient(
			deployURL,
			polytomic.APIKey(apiKey),
		)
	} else {
		p.client = polytomic.NewClient(
			deployURL,
			polytomic.DeploymentKey(deployKey),
		)
	}
	p.configured = true
}

func (p *ptProvider) GetResources(ctx context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return resources(), nil
}

func (p *ptProvider) GetDataSources(ctx context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{}, nil
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

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in provider.Provider) (ptProvider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*ptProvider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return ptProvider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return ptProvider{}, diags
	}

	return *p, diags
}
