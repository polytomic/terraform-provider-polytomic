// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SalesforceConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type SalesforceConnectionDataSource struct {
	provider *client.Provider
}

func (d *SalesforceConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *SalesforceConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_salesforce_connection"
}

func (d *SalesforceConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Salesforce Connection",
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
					"connect_mode": schema.StringAttribute{
						MarkdownDescription: "Default: browser (i.e. oauth through Polytomic). If 'code' is specified, the response will include an auth_code for the user to enter when completing authorization. NOTE: when supplying client_id and client_secret the connect mode must be 'api'.",
						Computed:            true,
					},
					"daily_api_calls": schema.Int64Attribute{
						MarkdownDescription: "The daily Salesforce API call cap that Polytomic should adhere to.",
						Computed:            true,
					},
					"domain": schema.StringAttribute{
						MarkdownDescription: "The Salesforce instance's login domain, e.g. acmecorp.my.salesforce.com",
						Computed:            true,
					},
					"enable_multicurrency_lookup": schema.BoolAttribute{
						MarkdownDescription: "If incremenetal mode for bulk-syncing from Salesforce formula fields is enabled, setting this to true extends support to accurate currency conversions.",
						Computed:            true,
					},
					"enable_tooling": schema.BoolAttribute{
						MarkdownDescription: "If true, expose objects from the Salesforce Tooling API in the Polytomic bulk sync source object list.",
						Computed:            true,
					},
					"enforce_api_limits": schema.BoolAttribute{
						MarkdownDescription: "If true, Polytomic will restrict itself to a fixed daily cap of Salesforce API calls enforced by the number in daily_api_calls.",
						Computed:            true,
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (d *SalesforceConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
			"connect_mode": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["connect_mode"], "string").(string),
			),
			"daily_api_calls": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["daily_api_calls"], "string").(string),
			),
			"domain": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["domain"], "string").(string),
			),
			"enable_multicurrency_lookup": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["enable_multicurrency_lookup"], "bool").(bool),
			),
			"enable_tooling": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["enable_tooling"], "bool").(bool),
			),
			"enforce_api_limits": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["enforce_api_limits"], "bool").(bool),
			),
			"username": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["username"], "string").(string),
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
