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
var _ datasource.DataSource = &QuickbooksConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type QuickbooksConnectionDataSource struct {
	provider *providerclient.Provider
}

func (d *QuickbooksConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *QuickbooksConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quickbooks_connection"
}

func (d *QuickbooksConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: QuickBooks Connection",
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
					"oauth_token_expiry": schema.StringAttribute{
						MarkdownDescription: ``,
						Computed:            true,
					},
					"realm_id": schema.StringAttribute{
						MarkdownDescription: `Company ID`,
						Computed:            true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (d *QuickbooksConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
			"oauth_token_expiry": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["oauth_token_expiry"], "string").(string),
			),
			"realm_id": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["realm_id"], "string").(string),
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
