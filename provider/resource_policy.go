package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
	ptcore "github.com/polytomic/polytomic-go/core"
	"github.com/polytomic/polytomic-go/permissions"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/client"
)

var _ resource.Resource = &policyResource{}
var _ resource.ResourceWithImportState = &policyResource{}

func (r *policyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Organizations: A policy in a Polytomic organization",
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"policy_actions": schema.SetNestedAttribute{
				MarkdownDescription: "Policy actions",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							MarkdownDescription: "Action",
							Required:            true,
						},
						"role_ids": schema.SetAttribute{
							MarkdownDescription: "Role IDs",
							Optional:            true,
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r policyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

type policyResourceData struct {
	Organization  types.String `tfsdk:"organization"`
	Name          types.String `tfsdk:"name"`
	Id            types.String `tfsdk:"id"`
	PolicyActions types.Set    `tfsdk:"policy_actions"`
}

type policyResource struct {
	provider *client.Provider
}

func (r *policyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data policyResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var policyActions []*polytomic.PolicyAction
	diags = data.PolicyActions.ElementsAs(ctx, &policyActions, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	policy, err := client.Permissions.Policies.Create(
		ctx,
		&permissions.CreatePolicyRequest{
			Name:           data.Name.ValueString(),
			OrganizationId: data.Organization.ValueStringPointer(),
			PolicyActions:  policyActions,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating policy: %s", err))
		return
	}

	tracked := make(map[string]bool)
	for _, action := range policyActions {
		tracked[action.Action] = true
	}
	// We only want to track the actions in the configuration
	// additional actions may be returned by the API
	// and we don't want to track them in the state
	var prunedPolicyActions []*polytomic.PolicyAction
	for _, action := range policy.Data.PolicyActions {
		if tracked[action.Action] {
			prunedPolicyActions = append(prunedPolicyActions, action)
			continue
		}

		if action.RoleIds != nil && len(action.RoleIds) > 0 {
			resp.Diagnostics.AddWarning(
				"Policy has actions not tracked by Terraform",
				fmt.Sprintf("Policy action %s has roles set but is not tracked in the state. This may cause data to be overwritten",
					strings.ToUpper(action.Action)),
			)
		}
	}
	resultPolicies, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"action":   types.StringType,
			"role_ids": types.SetType{ElemType: types.StringType},
		},
	}, prunedPolicyActions)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringPointerValue(policy.Data.Id)
	data.Name = types.StringPointerValue(policy.Data.Name)
	data.Organization = types.StringPointerValue(policy.Data.OrganizationId)
	data.PolicyActions = resultPolicies

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data policyResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var policyActions []polytomic.PolicyAction
	diags = data.PolicyActions.ElementsAs(ctx, &policyActions, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	policy, err := client.Permissions.Policies.Get(ctx, data.Id.ValueString())
	if err != nil {
		pErr := &ptcore.APIError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError("Error reading policy", err.Error())
		return
	}

	tracked := make(map[string]bool)
	for _, action := range policyActions {
		tracked[action.Action] = true
	}
	// We only want to track the actions in the configuration
	// additional actions may be returned by the API
	// and we don't want to track them in the state
	var prunedPolicyActions []*polytomic.PolicyAction
	for _, action := range policy.Data.PolicyActions {
		if tracked[action.Action] {
			prunedPolicyActions = append(prunedPolicyActions, action)
			continue
		}

		if action.RoleIds != nil && len(action.RoleIds) > 0 {
			resp.Diagnostics.AddWarning(
				"Policy has actions not tracked by Terraform",
				fmt.Sprintf("Policy action %s has roles set but is not tracked in the state. This may cause data to be overwritten",
					strings.ToUpper(action.Action)),
			)
		}
	}

	resultPolicies, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"action":   types.StringType,
			"role_ids": types.SetType{ElemType: types.StringType},
		},
	}, prunedPolicyActions)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringPointerValue(policy.Data.Id)
	data.Name = types.StringPointerValue(policy.Data.Name)
	data.Organization = types.StringPointerValue(policy.Data.OrganizationId)
	data.PolicyActions = resultPolicies

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data policyResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	var policyActions []*polytomic.PolicyAction
	diags = data.PolicyActions.ElementsAs(ctx, &policyActions, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	policy, err := client.Permissions.Policies.Update(
		ctx,
		data.Id.ValueString(),
		&permissions.UpdatePolicyRequest{
			Name:           data.Name.ValueString(),
			OrganizationId: data.Organization.ValueStringPointer(),
			PolicyActions:  policyActions,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating policy: %s", err))
		return
	}

	tracked := make(map[string]bool)
	for _, action := range policyActions {
		tracked[action.Action] = true
	}
	// We only want to track the actions in the configuration
	// additional actions may be returned by the API
	// and we don't want to track them in the state
	var prunedPolicyActions []*polytomic.PolicyAction
	for _, action := range policy.Data.PolicyActions {
		if tracked[action.Action] {
			prunedPolicyActions = append(prunedPolicyActions, action)
			continue
		}

		if action.RoleIds != nil && len(action.RoleIds) > 0 {
			resp.Diagnostics.AddWarning(
				"Policy has actions not tracked by Terraform",
				fmt.Sprintf("Policy action %s has roles set but is not tracked in the state. This may cause data to be overwritten",
					strings.ToUpper(action.Action)),
			)
		}
	}

	resultPolicies, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"action":   types.StringType,
			"role_ids": types.SetType{ElemType: types.StringType},
		},
	}, prunedPolicyActions)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringPointerValue(policy.Data.Id)
	data.Name = types.StringPointerValue(policy.Data.Name)
	data.Organization = types.StringPointerValue(policy.Data.OrganizationId)
	data.PolicyActions = resultPolicies

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data policyResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	err = client.Permissions.Policies.Remove(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting policy: %s", err))
		return
	}
}

func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
