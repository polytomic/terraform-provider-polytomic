package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
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
						}}},
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
	client *polytomic.Client
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data policyResourceData

	diags := req.Config.Get(ctx, &data)
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

	policy, err := r.client.Permissions().CreatePolicy(
		ctx,
		polytomic.PolicyRequest{
			Name:           data.Name.ValueString(),
			OrganizationID: data.Organization.ValueString(),
			PolicyActions:  policyActions,
		},
		polytomic.WithIdempotencyKey(uuid.NewString()),
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
	var prunedPolicyActions []polytomic.PolicyAction
	for _, action := range policy.PolicyActions {
		if tracked[action.Action] {
			prunedPolicyActions = append(prunedPolicyActions, action)
			continue
		}

		if action.RoleIDs != nil && len(action.RoleIDs) > 0 {
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

	data.Id = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Organization = types.StringValue(policy.OrganizationID)
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

	policy, err := r.client.Permissions().GetPolicy(ctx, data.Id.ValueString())
	if err != nil {
		pErr := polytomic.ApiError{}
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
	var prunedPolicyActions []polytomic.PolicyAction
	for _, action := range policy.PolicyActions {
		if tracked[action.Action] {
			prunedPolicyActions = append(prunedPolicyActions, action)
			continue
		}

		if action.RoleIDs != nil && len(action.RoleIDs) > 0 {
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

	data.Id = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Organization = types.StringValue(policy.OrganizationID)
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
	var policyActions []polytomic.PolicyAction
	diags = data.PolicyActions.ElementsAs(ctx, &policyActions, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	policy, err := r.client.Permissions().UpdatePolicy(
		ctx,
		data.Id.ValueString(),
		polytomic.PolicyRequest{
			Name:           data.Name.ValueString(),
			OrganizationID: data.Organization.ValueString(),
			PolicyActions:  policyActions,
		},
		polytomic.WithIdempotencyKey(uuid.NewString()),
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
	var prunedPolicyActions []polytomic.PolicyAction
	for _, action := range policy.PolicyActions {
		if tracked[action.Action] {
			prunedPolicyActions = append(prunedPolicyActions, action)
			continue
		}

		if action.RoleIDs != nil && len(action.RoleIDs) > 0 {
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

	data.Id = types.StringValue(policy.ID)
	data.Name = types.StringValue(policy.Name)
	data.Organization = types.StringValue(policy.OrganizationID)
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

	err := r.client.Permissions().DeletePolicy(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting policy: %s", err))
		return
	}
}

func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *policyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*polytomic.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}
