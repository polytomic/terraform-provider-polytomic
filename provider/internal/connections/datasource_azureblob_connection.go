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
var _ datasource.DataSource = &AzureblobConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type AzureblobConnectionDataSource struct {
	provider *providerclient.Provider
}

func (d *AzureblobConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *AzureblobConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azureblob_connection"
}

func (d *AzureblobConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Azure Blob Storage Connection",
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
					"account_name": schema.StringAttribute{
						MarkdownDescription: `Account Name`,
						Computed:            true,
					},
					"auth_method": schema.StringAttribute{
						MarkdownDescription: `Authentication method`,
						Computed:            true,
					},
					"client_id": schema.StringAttribute{
						MarkdownDescription: `Client ID`,
						Computed:            true,
					},
					"container_name": schema.StringAttribute{
						MarkdownDescription: `Container Name`,
						Computed:            true,
					},
					"is_single_table": schema.BoolAttribute{
						MarkdownDescription: `Files are time-based snapshots

    Treat the files as a single table.`,
						Computed: true,
					},
					"single_table_file_format": schema.StringAttribute{
						MarkdownDescription: `File format`,
						Computed:            true,
					},
					"single_table_name": schema.StringAttribute{
						MarkdownDescription: `Collection name`,
						Computed:            true,
					},
					"skip_lines": schema.Int64Attribute{
						MarkdownDescription: `Skip first lines

    Skip first N lines of each CSV file.`,
						Computed: true,
					},
					"tenant_id": schema.StringAttribute{
						MarkdownDescription: `Tenant ID`,
						Computed:            true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (d *AzureblobConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
			"account_name": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["account_name"], "string").(string),
			),
			"auth_method": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["auth_method"], "string").(string),
			),
			"client_id": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["client_id"], "string").(string),
			),
			"container_name": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["container_name"], "string").(string),
			),
			"is_single_table": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["is_single_table"], "bool").(bool),
			),
			"single_table_file_format": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["single_table_file_format"], "string").(string),
			),
			"single_table_name": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["single_table_name"], "string").(string),
			),
			"skip_lines": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["skip_lines"], "string").(string),
			),
			"tenant_id": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["tenant_id"], "string").(string),
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
