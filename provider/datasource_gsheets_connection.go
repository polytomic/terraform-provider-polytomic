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
var _ datasource.DataSource = &GsheetsConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type GsheetsConnectionDataSource struct {
	provider *client.Provider
}

func (d *GsheetsConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *GsheetsConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gsheets_connection"
}

func (d *GsheetsConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Google Sheets Connection",
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
						MarkdownDescription: "Default: browser",
						Computed:            true,
					},
					"has_headers": schema.BoolAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"oauth_token_expiry": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"spreadsheet_id": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"user_email": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (d *GsheetsConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
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
			"has_headers": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["has_headers"], "bool").(bool),
			),
			"oauth_token_expiry": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["oauth_token_expiry"], "string").(string),
			),
			"spreadsheet_id": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["spreadsheet_id"], "string").(string),
			),
			"user_email": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["user_email"], "string").(string),
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
