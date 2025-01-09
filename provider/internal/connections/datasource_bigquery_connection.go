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
var _ datasource.DataSource = &BigqueryConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type BigqueryConnectionDataSource struct {
	provider *providerclient.Provider
}

func (d *BigqueryConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *BigqueryConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_connection"
}

func (d *BigqueryConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Google BigQuery Connection",
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
					"client_email": schema.StringAttribute{
						MarkdownDescription: `Service account identity`,
						Computed:            true,
					},
					"location": schema.StringAttribute{
						MarkdownDescription: `Region or multi-region for query operations`,
						Computed:            true,
					},
					"override_project_id": schema.StringAttribute{
						MarkdownDescription: `Override project ID

    Override service key's project ID for cross-account access`,
						Computed: true,
					},
					"project_id": schema.StringAttribute{
						MarkdownDescription: `Service account project ID`,
						Computed:            true,
					},
					"structured_values_as_json": schema.BoolAttribute{
						MarkdownDescription: `Write object and array values as JSON`,
						Computed:            true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (d *BigqueryConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
			"client_email": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["client_email"], "string").(string),
			),
			"location": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["location"], "string").(string),
			),
			"override_project_id": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["override_project_id"], "string").(string),
			),
			"project_id": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["project_id"], "string").(string),
			),
			"structured_values_as_json": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["structured_values_as_json"], "bool").(bool),
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
