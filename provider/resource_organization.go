package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &organizationResource{}
var _ resource.ResourceWithImportState = &organizationResource{}

func (r *organizationResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}
func (r *organizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*polytomic.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *polytomic.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *organizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "polytomic_organization"
}

type organizationResourceData struct {
	Name      types.String `tfsdk:"name"`
	Id        types.String `tfsdk:"id"`
	SSODomain types.String `tfsdk:"sso_domain"`
	SSOOrgId  types.String `tfsdk:"sso_org_id"`
}

type organizationResource struct {
	client *polytomic.Client
}

func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data organizationResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Organizations().Create(ctx,
		polytomic.OrganizationMutation{
			Name:      data.Name.ValueString(),
			SSODomain: data.SSODomain.ValueString(),
			SSOOrgId:  data.SSOOrgId.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating organization: %s", err))
		return
	}
	data.Id = types.StringValue(created.ID.String())
	tflog.Trace(ctx, "created a organization")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data organizationResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	wsId, err := uuid.Parse(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Value Error", fmt.Sprintf("Invalid organization ID %s; error when parsing: %s", data.Id.ValueString(), err))
		return
	}
	organization, err := r.client.Organizations().Get(ctx, wsId)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading organization: %s", err))
		return
	}

	data.Id = types.StringValue(organization.ID.String())
	data.Name = types.StringValue(organization.Name)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data organizationResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	wsId, err := uuid.Parse(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Value Error", fmt.Sprintf("Invalid organization ID %s; error when parsing: %s", data.Id.ValueString(), err))
		return
	}

	updated, err := r.client.Organizations().Update(ctx, wsId,
		polytomic.OrganizationMutation{
			Name:      data.Name.ValueString(),
			SSODomain: data.SSODomain.ValueString(),
			SSOOrgId:  data.SSOOrgId.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating organization: %s", err))
		return
	}

	data.Name = types.StringValue(updated.Name)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TODO: Implement maybe?

	// We don't currently support deleting organizations via the API
	// since it's a destructive operation. We may want to support this
	// in the future, but for now we'll just log a warning.
	resp.Diagnostics.AddWarning("Deleting organizations is not currently supported by the API", "Please delete the organization manually.")
}

func (r *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
