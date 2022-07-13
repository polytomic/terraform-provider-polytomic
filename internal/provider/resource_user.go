package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.ResourceType = userResourceType{}
var _ tfsdk.Resource = userResource{}
var _ tfsdk.ResourceWithImportState = userResource{}

type userResourceType struct{}

func (t userResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "A User in a Polytomic Organization",

		Attributes: map[string]tfsdk.Attribute{
			"organization": {
				MarkdownDescription: "Organization ID",
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"email": {
				MarkdownDescription: "Email address",
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplaceIf(
						func(ctx context.Context, state, config attr.Value, path *tftypes.AttributePath) (bool, diag.Diagnostics) {
							tfState, err := state.ToTerraformValue(ctx)
							if err != nil {
								return false, nil
							}
							tfConfig, err := config.ToTerraformValue(ctx)
							if err != nil {
								return false, nil
							}

							var conf string
							var stt string

							if tfState.As(&stt) == nil && tfConfig.As(&conf) == nil {
								return !(strings.ToLower(stt) == strings.ToLower(conf)), nil
							}
							return false, nil
						},
						"Case-insensitively compares email addresses to determine if replacement is needed.",
						"Case-insensitively compares email addresses to determine if replacement is needed.",
					),
				},
			},
			"role": {
				MarkdownDescription: "Role; one of `user` or `admin`.",
				Optional:            true,
				Type:                types.StringType,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "user identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t userResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return userResource{
		provider: provider,
	}, diags
}

type userResourceData struct {
	Organization types.String `tfsdk:"organization"`
	Email        types.String `tfsdk:"email"`
	Role         types.String `tfsdk:"role"`
	Id           types.String `tfsdk:"id"`
}

type userResource struct {
	provider provider
}

func (r userResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data userResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.provider.client.Users().Create(ctx,
		uuid.MustParse(data.Organization.Value),
		polytomic.UserMutation{
			Email: data.Email.Value,
			Role:  data.Role.Value,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating user: %s", err))
		return
	}
	data.Id = types.String{Value: created.ID.String()}
	tflog.Trace(ctx, "created a user")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r userResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data userResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.provider.client.Users().Get(ctx, uuid.MustParse(data.Organization.Value), uuid.MustParse(data.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading user: %s", err))
		return
	}

	data.Id = types.String{Value: user.ID.String()}
	data.Organization = types.String{Value: user.OrganizationId.String()}
	data.Email = types.String{Value: user.Email}
	data.Role = types.String{Value: user.Role}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r userResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data userResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.provider.client.Users().Update(ctx,
		uuid.MustParse(data.Organization.Value),
		uuid.MustParse(data.Id.Value),
		polytomic.UserMutation{
			Email: data.Email.Value,
			Role:  data.Role.Value,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating user: %s", err))
		return
	}

	data.Id = types.String{Value: user.ID.String()}
	data.Organization = types.String{Value: user.OrganizationId.String()}
	data.Role = types.String{Value: user.Role}
	// Our backend normalizes email addresses to lowercase. As a result we do
	// not set the email here to prevent Terraform errors about inconsistent
	// state.

	// data.Email = types.String{Value: user.Email}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r userResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var data userResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.provider.client.Users().Delete(ctx, uuid.MustParse(data.Organization.Value), uuid.MustParse(data.Id.Value))
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting user: %s", err))
		return
	}
}

func (r userResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
