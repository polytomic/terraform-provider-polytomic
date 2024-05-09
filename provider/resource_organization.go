package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &organizationResource{}
var _ resource.ResourceWithImportState = &organizationResource{}

func (r *organizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Organizations: A Polytomic Organization provides a container for users, connections, and sync definitions.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Organization name",
				Required:            true,
			},
			"sso_domain": schema.StringAttribute{
				MarkdownDescription: "Single sign-on domain",
				Optional:            true,
			},
			"sso_org_id": schema.StringAttribute{
				MarkdownDescription: "Single sign-on organization ID (WorkOS)",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Organization identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
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
		polytomic.WithIdempotencyKey(uuid.NewString()),
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
		pErr := polytomic.ApiError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
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
		polytomic.WithIdempotencyKey(uuid.NewString()),
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

	err = r.client.Organizations().Delete(ctx, wsId)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting organization", err.Error())
		return
	}
}

func (r *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
