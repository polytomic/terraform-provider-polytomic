package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
	ptclient "github.com/polytomic/polytomic-go/client"
	ptcore "github.com/polytomic/polytomic-go/core"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &userResource{}
var _ resource.ResourceWithImportState = &userResource{}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Organizations: A user in a Polytomic organization",
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
						resp.RequiresReplace = !strings.EqualFold(req.StateValue.ValueString(), req.ConfigValue.ValueString())
					},
						"Case-insensitively compares email addresses to determine if replacement is needed.",
						"Case-insensitively compares email addresses to determine if replacement is needed.",
					),
				},
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role; one of `user` or `admin`.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "user identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

type userResourceData struct {
	Organization types.String `tfsdk:"organization"`
	Email        types.String `tfsdk:"email"`
	Role         types.String `tfsdk:"role"`
	Id           types.String `tfsdk:"id"`
}

type userResource struct {
	client *ptclient.Client
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Users.Create(ctx,
		data.Organization.ValueString(),
		&polytomic.CreateUserRequestSchema{
			Email: data.Email.ValueString(),
			Role:  data.Role.ValueStringPointer(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating user: %s", err))
		return
	}
	data.Id = types.StringPointerValue(created.Data.Id)
	tflog.Trace(ctx, "created a user")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.Users.Get(ctx, data.Id.ValueString(), data.Organization.ValueString())
	if err != nil {
		pErr := &ptcore.APIError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading user: %s", err))
		return
	}
	if user.Data.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Id = types.StringPointerValue(user.Data.Id)
	data.Organization = types.StringPointerValue(user.Data.OrganizationId)
	data.Role = types.StringPointerValue(user.Data.Role)

	// Our backend normalizes email addresses to lowercase. As a result,
	// we need to do the same here to ensure that the state is consistent
	if !strings.EqualFold(data.Email.ValueString(), pointer.GetString(user.Data.Email)) {
		data.Email = types.StringPointerValue(user.Data.Email)
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data userResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.Users.Update(ctx,
		data.Id.ValueString(),
		data.Organization.ValueString(),
		&polytomic.UpdateUserRequestSchema{
			Email: data.Email.ValueString(),
			Role:  data.Role.ValueStringPointer(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating user: %s", err))
		return
	}

	data.Id = types.StringPointerValue(user.Data.Id)
	data.Organization = types.StringPointerValue(user.Data.OrganizationId)
	data.Role = types.StringPointerValue(user.Data.Role)
	// Our backend normalizes email addresses to lowercase. As a result we do
	// not set the email here to prevent Terraform errors about inconsistent
	// state.

	// data.Email = types.String{Value: user.Email}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Users.Remove(ctx, data.Id.ValueString(), data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting user: %s", err))
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ptclient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}
