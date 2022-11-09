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
var _ resource.Resource = &sqlserverConnectionResource{}
var _ resource.ResourceWithImportState = &sqlserverConnectionResource{}

func (t *sqlserverConnectionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "SQL Server Connection",
		Attributes: map[string]tfsdk.Attribute{
			"organization": {
				MarkdownDescription: "Organization ID",
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
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
				}),
				Required: true,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "SQL Server Connection identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (r *sqlserverConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "polytomic_sqlserver_connection"
}

type sqlserverConnectionResource struct {
	client *polytomic.Client
}

func (r *sqlserverConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:           data.Name.ValueString(),
			Type:           polytomic.SQLServerConnectionType,
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.SQLServerConfiguration{
				Hostname: data.Configuration.Attributes()["hostname"].(types.String).ValueString(),
				Username: data.Configuration.Attributes()["username"].(types.String).ValueString(),
				Password: data.Configuration.Attributes()["password"].(types.String).ValueString(),
				Database: data.Configuration.Attributes()["database"].(types.String).ValueString(),
				Port:     int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}
	data.Id = types.StringValue(created.ID)
	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "sqlserver", "id": created.ID})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *sqlserverConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionResourceData

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
	data.Organization = types.StringValue(connection.OrganizationId)
	data.Name = types.StringValue(connection.Name)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *sqlserverConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data connectionResourceData

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
			Configuration: polytomic.SQLServerConfiguration{
				Hostname: data.Configuration.Attributes()["hostname"].(types.String).ValueString(),
				Username: data.Configuration.Attributes()["username"].(types.String).ValueString(),
				Password: data.Configuration.Attributes()["password"].(types.String).ValueString(),
				Database: data.Configuration.Attributes()["database"].(types.String).ValueString(),
				Port:     int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	data.Id = types.StringValue(updated.ID)
	data.Organization = types.StringValue(updated.OrganizationId)
	data.Name = types.StringValue(updated.Name)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *sqlserverConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionResourceData

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

func (r *sqlserverConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *sqlserverConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
