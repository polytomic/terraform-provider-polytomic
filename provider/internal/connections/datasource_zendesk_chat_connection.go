// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package connections

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/providerclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &Zendesk_chatConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type Zendesk_chatConnectionDataSource struct {
	provider *providerclient.Provider
}

func (d *Zendesk_chatConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *Zendesk_chatConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zendesk_chat_connection"
}

func (d *Zendesk_chatConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Zendesk Chat Connection",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"configuration": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"custom_api_limits": schema.BoolAttribute{
						MarkdownDescription: `Enforce custom API limits`,
						Computed:            true,
					},
					"domain": schema.StringAttribute{
						MarkdownDescription: `Zendesk Support Subdomain`,
						Computed:            true,
					},
					"ratelimit_rpm": schema.Int64Attribute{
						MarkdownDescription: `Maximum requests per minute

    Set a custom maximum request per minute limit.`,
						Computed: true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (d *Zendesk_chatConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data connectionData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the connection
	client, err := d.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	connection, err := client.Connections.Get(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection", err.Error())
		return
	}

	data.Id = types.StringPointerValue(connection.Data.Id)
	data.Name = types.StringPointerValue(connection.Data.Name)
	data.Organization = types.StringPointerValue(connection.Data.OrganizationId)
	var diags diag.Diagnostics
	data.Configuration, diags = types.ObjectValue(
		data.Configuration.AttributeTypes(ctx),
		map[string]attr.Value{
			"custom_api_limits": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["custom_api_limits"], "bool").(bool),
			),
			"domain": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["domain"], "string").(string),
			),
			"ratelimit_rpm": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["ratelimit_rpm"], "string").(string),
			),
		},
	)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
