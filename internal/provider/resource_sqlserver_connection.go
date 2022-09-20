package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.ResourceType = sqlserverConnectionResourceType{}
var _ resource.Resource = sqlserverConnectionResource{}
var _ resource.ResourceWithImportState = sqlserverConnectionResource{}

type sqlserverConnectionResourceType struct{}

func (t sqlserverConnectionResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
						Type:     types.StringType,
						Required: true,
					},
					"username": {
						Type:     types.StringType,
						Required: true,
					},
					"password": {
						Type:     types.StringType,
						Required: true,
					},
					"database": {
						Type:     types.StringType,
						Required: true,
					},
					"port": {
						Type:     types.Int64Type,
						Required: true,
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

func (t sqlserverConnectionResourceType) NewResource(ctx context.Context, in provider.Provider) (resource.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return sqlserverConnectionResource{
		provider: provider,
	}, diags
}

type sqlserverConnectionResource struct {
	provider ptProvider
}

func (r sqlserverConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.provider.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:           data.Name.Value,
			Type:           polytomic.SQLServerConnectionType,
			OrganizationId: data.Organization.Value,
			Configuration: polytomic.SQLServerConfiguration{
				Hostname: data.Configuration.Attrs["hostname"].(types.String).Value,
				Username: data.Configuration.Attrs["username"].(types.String).Value,
				Password: data.Configuration.Attrs["password"].(types.String).Value,
				Database: data.Configuration.Attrs["database"].(types.String).Value,
				Port:     int(data.Configuration.Attrs["port"].(types.Int64).Value),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}
	data.Id = types.String{Value: created.ID}
	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "sqlserver"})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r sqlserverConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	connection, err := r.provider.client.Connections().Get(ctx, uuid.MustParse(data.Id.Value))
	if err != nil {
		if err.Error() == ConnectionNotFoundErr {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading connection: %s", err))
		return
	}

	data.Id = types.String{Value: connection.ID}
	data.Organization = types.String{Value: connection.OrganizationId}
	data.Name = types.String{Value: connection.Name}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r sqlserverConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data connectionResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.provider.client.Connections().Update(ctx,
		uuid.MustParse(data.Id.Value),
		polytomic.UpdateConnectionMutation{
			Name:           data.Name.Value,
			OrganizationId: data.Organization.Value,
			Configuration: polytomic.SQLServerConfiguration{
				Hostname: data.Configuration.Attrs["hostname"].(types.String).Value,
				Username: data.Configuration.Attrs["username"].(types.String).Value,
				Password: data.Configuration.Attrs["password"].(types.String).Value,
				Database: data.Configuration.Attrs["database"].(types.String).Value,
				Port:     int(data.Configuration.Attrs["port"].(types.Int64).Value),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	data.Id = types.String{Value: updated.ID}
	data.Organization = types.String{Value: updated.OrganizationId}
	data.Name = types.String{Value: updated.Name}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r sqlserverConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.provider.client.Connections().Delete(ctx, uuid.MustParse(data.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting connection: %s", err))
		return
	}
}

func (r sqlserverConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
