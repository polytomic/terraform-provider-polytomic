// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &PostgresqlConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type PostgresqlConnectionDataSource struct {
	client *polytomic.Client
}

func (d *PostgresqlConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql_connection"
}

func (d *PostgresqlConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: PostgresSQL Connection",
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
					"hostname": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"database": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"ssl": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"client_certs": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"client_certificate": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"ca_cert": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"change_detection": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"publication": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"ssh": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"ssh_user": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"ssh_host": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"ssh_port": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
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

func (d *PostgresqlConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*polytomic.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *polytomic.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *PostgresqlConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data connectionData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the connection
	connection, err := d.client.Connections().Get(ctx, uuid.MustParse(data.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection", err.Error())
		return
	}

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue(connection.ID)
	data.Name = types.StringValue(connection.Name)
	data.Organization = types.StringValue(connection.OrganizationId)
	var conf polytomic.PostgresqlConfiguration
	err = mapstructure.Decode(connection.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError("Error decoding connection", err.Error())
		return
	}

	var diags diag.Diagnostics
	data.Configuration, diags = types.ObjectValue(
		data.Configuration.AttributeTypes(ctx),
		map[string]attr.Value{
			"hostname": types.StringValue(
				conf.Hostname,
			),
			"username": types.StringValue(
				conf.Username,
			),
			"database": types.StringValue(
				conf.Database,
			),
			"port": types.Int64Value(
				int64(conf.Port),
			),
			"ssl": types.BoolValue(
				conf.SSL,
			),
			"client_certs": types.BoolValue(
				conf.ClientCerts,
			),
			"client_certificate": types.StringValue(
				conf.ClientCertificate,
			),
			"ca_cert": types.StringValue(
				conf.CACert,
			),
			"change_detection": types.BoolValue(
				conf.ChangeDetection,
			),
			"publication": types.StringValue(
				conf.Publication,
			),
			"ssh": types.BoolValue(
				conf.SSH,
			),
			"ssh_user": types.StringValue(
				conf.SSHUser,
			),
			"ssh_host": types.StringValue(
				conf.SSHHost,
			),
			"ssh_port": types.Int64Value(
				int64(conf.SSHPort),
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
