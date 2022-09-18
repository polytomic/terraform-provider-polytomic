package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
var _ provider.ResourceType = userResourceType{}
var _ resource.Resource = userResource{}
var _ resource.ResourceWithImportState = userResource{}

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
					resource.RequiresReplace(),
				},
			},
			"email": {
				MarkdownDescription: "Email address",
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplaceIf(
						func(ctx context.Context, state, config attr.Value, path path.Path) (bool, diag.Diagnostics) {
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
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (t userResourceType) NewResource(ctx context.Context, in provider.Provider) (resource.Resource, diag.Diagnostics) {
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
	provider ptProvider
}

func (r userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

func (r userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	data.Role = types.String{Value: user.Role}

	// Our backend normalizes email addresses to lowercase. As a result,
	// we need to do the same here to ensure that the state is consistent
	if strings.ToLower(data.Email.Value) != strings.ToLower(user.Email) {
		data.Email = types.String{Value: user.Email}
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

func (r userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
