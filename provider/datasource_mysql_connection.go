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
var _ datasource.DataSource = &MysqlConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type MysqlConnectionDataSource struct {
	provider *client.Provider
}

func (d *MysqlConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *MysqlConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_connection"
}

func (d *MysqlConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: MySQL Connection",
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
					"account": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"change_detection": schema.BoolAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"dbname": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"hostname": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"ssh": schema.BoolAttribute{
						MarkdownDescription: "",
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
					"ssl": schema.BoolAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (d *MysqlConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
			"account": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["account"], "string").(string),
			),
			"change_detection": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["change_detection"], "bool").(bool),
			),
			"dbname": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["dbname"], "string").(string),
			),
			"hostname": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["hostname"], "string").(string),
			),
			"port": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["port"], "string").(string),
			),
			"ssh": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["ssh"], "bool").(bool),
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
			"ssl": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["ssl"], "bool").(bool),
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
