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
var _ datasource.DataSource = &PostgresqlConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type PostgresqlConnectionDataSource struct {
	provider *client.Provider
}

func (d *PostgresqlConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *PostgresqlConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql_connection"
}

func (d *PostgresqlConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: PostgreSQL Connection",
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
					"change_detection": schema.BoolAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"client_certs": schema.BoolAttribute{
						MarkdownDescription: "",
						Computed:            true,
					},
					"database": schema.StringAttribute{
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
					"publication": schema.StringAttribute{
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

func (d *PostgresqlConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
			"change_detection": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["change_detection"], "bool").(bool),
			),
			"client_certs": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["client_certs"], "bool").(bool),
			),
			"database": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["database"], "string").(string),
			),
			"hostname": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["hostname"], "string").(string),
			),
			"port": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["port"], "string").(string),
			),
			"publication": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["publication"], "string").(string),
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
