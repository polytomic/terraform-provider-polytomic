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
	ptcore "github.com/polytomic/polytomic-go/core"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
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

type userResourceData struct {
	Organization types.String `tfsdk:"organization"`
	Email        types.String `tfsdk:"email"`
	Role         types.String `tfsdk:"role"`
	Id           types.String `tfsdk:"id"`
}

type userResource struct {
	provider *providerclient.Provider
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	created, err := client.Users.Create(ctx,
		data.Organization.ValueString(),
		&polytomic.CreateUserRequestSchema{
			Email: data.Email.ValueString(),
			Role:  data.Role.ValueStringPointer(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error creating user: %s", err))
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

	orgID := data.Organization.ValueString()
	client, err := r.provider.Client(orgID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	user, err := client.Users.Get(ctx, data.Id.ValueString(), orgID)
	if err != nil {
		pErr := &ptcore.APIError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error reading user: %s", err))
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

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	user, err := client.Users.Update(ctx,
		data.Id.ValueString(),
		data.Organization.ValueString(),
		&polytomic.UpdateUserRequestSchema{
			Email: data.Email.ValueString(),
			Role:  data.Role.ValueStringPointer(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error updating user: %s", err))
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
	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	_, err = client.Users.Remove(ctx, data.Id.ValueString(), data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error deleting user: %s", err))
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var organizationID, identifier string

	// Parse import ID format
	parts := strings.Split(req.ID, "/")
	if len(parts) == 2 {
		// Compound format: org_id/identifier (identifier can be user_id or email)
		organizationID = parts[0]
		identifier = parts[1]
	} else if len(parts) == 1 {
		// Simple format: just identifier, auto-detect organization from caller identity
		identifier = req.ID

		client, err := r.provider.Client("")
		if err != nil {
			resp.Diagnostics.AddError("Error getting client", err.Error())
			return
		}

		identity, err := client.Identity.Get(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Error getting caller identity", fmt.Sprintf("Unable to determine organization from API key: %s", err))
			return
		}

		if identity.Data.OrganizationId == nil {
			resp.Diagnostics.AddError("Error getting organization", "Caller identity does not have an organization ID")
			return
		}

		organizationID = *identity.Data.OrganizationId
	} else {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected import ID in format 'identifier' or 'org_id/identifier' where identifier is a user ID or email address, got: %s", req.ID),
		)
		return
	}

	// Determine if identifier is an email address or user ID
	var userID string
	if strings.Contains(identifier, "@") {
		// Identifier is an email address - look up the user ID
		tflog.Debug(ctx, "Importing user by email address", map[string]any{
			"org_id": organizationID,
			"email":  identifier,
		})

		client, err := r.provider.Client(organizationID)
		if err != nil {
			resp.Diagnostics.AddError("Error getting client", err.Error())
			return
		}

		// List all users in the organization
		users, err := client.Users.List(ctx, organizationID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error listing users",
				fmt.Sprintf("Failed to list users in organization %s: %s", organizationID, err),
			)
			return
		}

		// Find user by email (case-insensitive comparison)
		var foundUser *polytomic.User
		for _, user := range users.Data {
			if user.Email != nil && strings.EqualFold(*user.Email, identifier) {
				foundUser = user
				break
			}
		}

		if foundUser == nil {
			resp.Diagnostics.AddError(
				"User not found",
				fmt.Sprintf("No user with email address %s found in organization %s", identifier, organizationID),
			)
			return
		}

		if foundUser.Id == nil {
			resp.Diagnostics.AddError(
				"Invalid user data",
				fmt.Sprintf("User with email %s has no ID", identifier),
			)
			return
		}

		userID = *foundUser.Id
		tflog.Debug(ctx, "Found user by email", map[string]any{
			"email":   identifier,
			"user_id": userID,
		})
	} else {
		// Identifier is a user ID
		userID = identifier
		tflog.Debug(ctx, "Importing user by ID", map[string]any{
			"org_id":  organizationID,
			"user_id": userID,
		})
	}

	// Set both the ID and organization in the state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), userID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), organizationID)...)
}
