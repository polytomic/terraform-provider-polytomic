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
var _ datasource.DataSource = &SftpConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type SftpConnectionDataSource struct {
	provider *client.Provider
}

func (d *SftpConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *SftpConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sftp_connection"
}

func (d *SftpConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: SFTP Connection",
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
					"is_single_table": schema.BoolAttribute{
						MarkdownDescription: "Treat the files as a single table.",
						Computed:            true,
					},
					"path": schema.StringAttribute{
						MarkdownDescription: "The path to the directory on the SFTP server containing the files.",
						Computed:            true,
					},
					"single_table_name": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"skip_lines": schema.Int64Attribute{
						MarkdownDescription: "Skip first N lines of each CSV file.",
						Computed:            true,
					},
					"ssh_host": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"ssh_port": schema.Int64Attribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"ssh_user": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (d *SftpConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
			"is_single_table": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["is_single_table"], "bool").(bool),
			),
			"path": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["path"], "string").(string),
			),
			"single_table_name": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["single_table_name"], "string").(string),
			),
			"skip_lines": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["skip_lines"], "string").(string),
			),
			"ssh_host": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["ssh_host"], "string").(string),
			),
			"ssh_port": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["ssh_port"], "string").(string),
			),
			"ssh_user": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["ssh_user"], "string").(string),
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