package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

const clientError = "Client Error"

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = workspaceResourceType{}
var _ tfsdk.Resource = workspaceResource{}
var _ tfsdk.ResourceWithImportState = workspaceResource{}

type workspaceResourceType struct{}

func (t workspaceResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "A Polytomic Workspace provides a container for users, connections, and sync definitions.",

		Attributes: map[string]tfsdk.Attribute{
			"name": {
				MarkdownDescription: "Workspace name",
				Required:            true,
				Type:                types.StringType,
			},
			"sso_domain": {
				MarkdownDescription: "Single sign-on domain",
				Optional:            true,
				Type:                types.StringType,
			},
			"sso_org_id": {
				MarkdownDescription: "Single sign-on organization ID (WorkOS)",
				Optional:            true,
				Type:                types.StringType,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Workspace identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t workspaceResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return workspaceResource{
		provider: provider,
	}, diags
}

type workspaceResourceData struct {
	Name      types.String `tfsdk:"name"`
	Id        types.String `tfsdk:"id"`
	SSODomain types.String `tfsdk:"sso_domain"`
	SSOOrgId  types.String `tfsdk:"sso_org_id"`
}

type workspaceResource struct {
	provider provider
}

func (r workspaceResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data workspaceResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.provider.client.Workspaces().Create(ctx,
		polytomic.WorkspaceMutation{
			Name:      data.Name.Value,
			SSODomain: data.SSODomain.Value,
			SSOOrgId:  data.SSOOrgId.Value,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating workspace: %s", err))
		return
	}
	data.Id = types.String{Value: created.ID.String()}
	tflog.Trace(ctx, "created a workspace")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r workspaceResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data workspaceResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	wsId, err := uuid.Parse(data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Value Error", fmt.Sprintf("Invalid workspace ID %s; error when parsing: %s", data.Id.Value, err))
		return
	}
	workspace, err := r.provider.client.Workspaces().Get(ctx, wsId)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading workspace: %s", err))
		return
	}

	data.Id = types.String{Value: workspace.ID.String()}
	data.Name = types.String{Value: workspace.Name}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r workspaceResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data workspaceResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	wsId, err := uuid.Parse(data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Value Error", fmt.Sprintf("Invalid workspace ID %s; error when parsing: %s", data.Id.Value, err))
		return
	}

	updated, err := r.provider.client.Workspaces().Update(ctx, wsId,
		polytomic.WorkspaceMutation{
			Name:      data.Name.Value,
			SSODomain: data.SSODomain.Value,
			SSOOrgId:  data.SSOOrgId.Value,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating workspace: %s", err))
		return
	}

	data.Name = types.String{Value: updated.Name}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r workspaceResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data workspaceResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	wsId, err := uuid.Parse(data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Value Error", fmt.Sprintf("Invalid workspace ID %s; error when parsing: %s", data.Id.Value, err))
		return
	}
	err = r.provider.client.Workspaces().Delete(ctx, wsId)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting workspace: %s", err))
		return
	}
}

func (r workspaceResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
