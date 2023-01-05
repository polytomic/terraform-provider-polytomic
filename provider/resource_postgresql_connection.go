// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &PostgresqlConnectionResource{}
var _ resource.ResourceWithImportState = &PostgresqlConnectionResource{}

func (t *PostgresqlConnectionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "PostgresSQL Connection",
		Attributes: map[string]tfsdk.Attribute{
			"organization": {
				MarkdownDescription: "Organization ID",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"configuration": {
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"hostname": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"username": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"password": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           true,
					},
					"database": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"port": {
						MarkdownDescription: "",
						Type:                types.Int64Type,
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"ssl": {
						MarkdownDescription: "",
						Type:                types.BoolType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"client_certs": {
						MarkdownDescription: "",
						Type:                types.BoolType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"client_certificate": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"client_key": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           true,
					},
					"ca_cert": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"change_detection": {
						MarkdownDescription: "",
						Type:                types.BoolType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"publication": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh": {
						MarkdownDescription: "",
						Type:                types.BoolType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh_user": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh_host": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh_port": {
						MarkdownDescription: "",
						Type:                types.Int64Type,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"private_key": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           true,
					},
				}),

				Required: true,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "PostgresSQL Connection identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (r *PostgresqlConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql_connection"
}

type PostgresqlConnectionResource struct {
	client *polytomic.Client
}

func (r *PostgresqlConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:           data.Name.ValueString(),
			Type:           polytomic.PostgresqlConnectionType,
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.PostgresqlConfiguration{
				Hostname:          data.Configuration.Attributes()["hostname"].(types.String).ValueString(),
				Username:          data.Configuration.Attributes()["username"].(types.String).ValueString(),
				Password:          data.Configuration.Attributes()["password"].(types.String).ValueString(),
				Database:          data.Configuration.Attributes()["database"].(types.String).ValueString(),
				Port:              int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
				SSL:               data.Configuration.Attributes()["ssl"].(types.Bool).ValueBool(),
				ClientCerts:       data.Configuration.Attributes()["client_certs"].(types.Bool).ValueBool(),
				ClientCertificate: data.Configuration.Attributes()["client_certificate"].(types.String).ValueString(),
				ClientKey:         data.Configuration.Attributes()["client_key"].(types.String).ValueString(),
				CACert:            data.Configuration.Attributes()["ca_cert"].(types.String).ValueString(),
				ChangeDetection:   data.Configuration.Attributes()["change_detection"].(types.Bool).ValueBool(),
				Publication:       data.Configuration.Attributes()["publication"].(types.String).ValueString(),
				SSH:               data.Configuration.Attributes()["ssh"].(types.Bool).ValueBool(),
				SSHUser:           data.Configuration.Attributes()["ssh_user"].(types.String).ValueString(),
				SSHHost:           data.Configuration.Attributes()["ssh_host"].(types.String).ValueString(),
				SSHPort:           int(data.Configuration.Attributes()["ssh_port"].(types.Int64).ValueInt64()),
				PrivateKey:        data.Configuration.Attributes()["private_key"].(types.String).ValueString(),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}
	data.Id = types.StringValue(created.ID)
	data.Name = types.StringValue(created.Name)

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Postgresql", "id": created.ID})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresqlConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	connection, err := r.client.Connections().Get(ctx, uuid.MustParse(data.Id.ValueString()))
	if err != nil {
		if err.Error() == ConnectionNotFoundErr {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading connection: %s", err))
		return
	}

	data.Id = types.StringValue(connection.ID)
	data.Name = types.StringValue(connection.Name)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresqlConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data connectionData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.Connections().Update(ctx,
		uuid.MustParse(data.Id.ValueString()),
		polytomic.UpdateConnectionMutation{
			Name:           data.Name.ValueString(),
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.PostgresqlConfiguration{
				Hostname:          data.Configuration.Attributes()["hostname"].(types.String).ValueString(),
				Username:          data.Configuration.Attributes()["username"].(types.String).ValueString(),
				Password:          data.Configuration.Attributes()["password"].(types.String).ValueString(),
				Database:          data.Configuration.Attributes()["database"].(types.String).ValueString(),
				Port:              int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
				SSL:               data.Configuration.Attributes()["ssl"].(types.Bool).ValueBool(),
				ClientCerts:       data.Configuration.Attributes()["client_certs"].(types.Bool).ValueBool(),
				ClientCertificate: data.Configuration.Attributes()["client_certificate"].(types.String).ValueString(),
				ClientKey:         data.Configuration.Attributes()["client_key"].(types.String).ValueString(),
				CACert:            data.Configuration.Attributes()["ca_cert"].(types.String).ValueString(),
				ChangeDetection:   data.Configuration.Attributes()["change_detection"].(types.Bool).ValueBool(),
				Publication:       data.Configuration.Attributes()["publication"].(types.String).ValueString(),
				SSH:               data.Configuration.Attributes()["ssh"].(types.Bool).ValueBool(),
				SSHUser:           data.Configuration.Attributes()["ssh_user"].(types.String).ValueString(),
				SSHHost:           data.Configuration.Attributes()["ssh_host"].(types.String).ValueString(),
				SSHPort:           int(data.Configuration.Attributes()["ssh_port"].(types.Int64).ValueInt64()),
				PrivateKey:        data.Configuration.Attributes()["private_key"].(types.String).ValueString(),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	data.Id = types.StringValue(updated.ID)
	data.Name = types.StringValue(updated.Name)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresqlConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Connections().Delete(ctx, uuid.MustParse(data.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting connection: %s", err))
		return
	}
}

func (r *PostgresqlConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PostgresqlConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}
