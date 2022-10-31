// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

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
var _ provider.ResourceType = bigqueryConnectionResourceType{}
var _ resource.Resource = bigqueryConnectionResource{}
var _ resource.ResourceWithImportState = bigqueryConnectionResource{}

type bigqueryConnectionResourceType struct{}

func (t bigqueryConnectionResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Big Query Connection",
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
					"service_account_credentials": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           true,
					},
					"location": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
				}),
				Required: true,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Big Query Connection identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t bigqueryConnectionResourceType) NewResource(ctx context.Context, in provider.Provider) (resource.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return bigqueryConnectionResource{
		provider: provider,
	}, diags
}

type bigqueryConnectionResource struct {
	provider ptProvider
}

func (r bigqueryConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.provider.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:           data.Name.Value,
			Type:           polytomic.BigQueryConnectionType,
			OrganizationId: data.Organization.Value,
			Configuration: polytomic.BigQueryConfiguration{
				ServiceAccountCredentials: data.Configuration.Attrs["service_account_credentials"].(types.String).Value,
				Location:                  data.Configuration.Attrs["location"].(types.String).Value,
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}
	data.Id = types.String{Value: created.ID}
	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "bigquery", "id": created.ID})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r bigqueryConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

func (r bigqueryConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
			Configuration: polytomic.BigQueryConfiguration{
				ServiceAccountCredentials: data.Configuration.Attrs["service_account_credentials"].(types.String).Value,
				Location:                  data.Configuration.Attrs["location"].(types.String).Value,
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

func (r bigqueryConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r bigqueryConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
