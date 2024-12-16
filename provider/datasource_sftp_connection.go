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
			"name": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"configuration": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"is_single_table": schema.BoolAttribute{
						MarkdownDescription: "Treat the files as a single table.",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
					"path": schema.StringAttribute{
						MarkdownDescription: "The path to the directory on the SFTP server containing the files.",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
					"single_table_name": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
					"skip_lines": schema.Int64Attribute{
						MarkdownDescription: "Skip first N lines of each CSV file.",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
					"ssh_host": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
					"ssh_port": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
					"ssh_private_key": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
					"ssh_user": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
				},
				Optional: true,
			},
			"force_destroy": schema.BoolAttribute{
				MarkdownDescription: forceDestroyMessage,
				Optional:            true,
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
			"ssh_private_key": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["ssh_private_key"], "string").(string),
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
