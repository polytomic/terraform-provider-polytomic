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
var _ tfsdk.ResourceType = organizationResourceType{}
var _ tfsdk.Resource = organizationResource{}
var _ tfsdk.ResourceWithImportState = organizationResource{}

type organizationResourceType struct{}

func (t organizationResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "A Polytomic Organization provides a container for users, connections, and sync definitions.",

		Attributes: map[string]tfsdk.Attribute{
			"name": {
				MarkdownDescription: "Organization name",
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
				MarkdownDescription: "Organization identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t organizationResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return organizationResource{
		provider: provider,
	}, diags
}

type organizationResourceData struct {
	Name      types.String `tfsdk:"name"`
	Id        types.String `tfsdk:"id"`
	SSODomain types.String `tfsdk:"sso_domain"`
	SSOOrgId  types.String `tfsdk:"sso_org_id"`
}

type organizationResource struct {
	provider provider
}

func (r organizationResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data organizationResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.provider.client.Organizations().Create(ctx,
		polytomic.OrganizationMutation{
			Name:      data.Name.Value,
			SSODomain: data.SSODomain.Value,
			SSOOrgId:  data.SSOOrgId.Value,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating organization: %s", err))
		return
	}
	data.Id = types.String{Value: created.ID.String()}
	tflog.Trace(ctx, "created a organization")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r organizationResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data organizationResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	wsId, err := uuid.Parse(data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Value Error", fmt.Sprintf("Invalid organization ID %s; error when parsing: %s", data.Id.Value, err))
		return
	}
	organization, err := r.provider.client.Organizations().Get(ctx, wsId)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading organization: %s", err))
		return
	}

	data.Id = types.String{Value: organization.ID.String()}
	data.Name = types.String{Value: organization.Name}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r organizationResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data organizationResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	wsId, err := uuid.Parse(data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Value Error", fmt.Sprintf("Invalid organization ID %s; error when parsing: %s", data.Id.Value, err))
		return
	}

	updated, err := r.provider.client.Organizations().Update(ctx, wsId,
		polytomic.OrganizationMutation{
			Name:      data.Name.Value,
			SSODomain: data.SSODomain.Value,
			SSOOrgId:  data.SSOOrgId.Value,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating organization: %s", err))
		return
	}

	data.Name = types.String{Value: updated.Name}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r organizationResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data organizationResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	wsId, err := uuid.Parse(data.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError("Value Error", fmt.Sprintf("Invalid organization ID %s; error when parsing: %s", data.Id.Value, err))
		return
	}
	err = r.provider.client.Organizations().Delete(ctx, wsId)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting organization: %s", err))
		return
	}
}

func (r organizationResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
