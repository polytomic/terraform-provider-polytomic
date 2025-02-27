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
var _ datasource.DataSource = &GcsConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type GcsConnectionDataSource struct {
	provider *providerclient.Provider
}

func (d *GcsConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *GcsConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcs_connection"
}

func (d *GcsConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Google Cloud Storage Connection",
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
					"bucket": schema.StringAttribute{
						MarkdownDescription: ``,
						Computed:            true,
					},
					"client_email": schema.StringAttribute{
						MarkdownDescription: `Service account identity`,
						Computed:            true,
					},
					"is_single_table": schema.BoolAttribute{
						MarkdownDescription: `Files are time-based snapshots

    Treat the files as a single table.`,
						Computed: true,
					},
					"project_id": schema.StringAttribute{
						MarkdownDescription: `Service account project ID`,
						Computed:            true,
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
				},
				Optional: true,
			},
		},
	}
}

func (d *GcsConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
			"bucket": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["bucket"], "string").(string),
			),
			"client_email": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["client_email"], "string").(string),
			),
			"is_single_table": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["is_single_table"], "bool").(bool),
			),
			"project_id": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["project_id"], "string").(string),
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
		},
	)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
