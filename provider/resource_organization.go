package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/client"
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

type organizationResourceData struct {
	Name      types.String `tfsdk:"name"`
	Id        types.String `tfsdk:"id"`
	SSODomain types.String `tfsdk:"sso_domain"`
	SSOOrgId  types.String `tfsdk:"sso_org_id"`
}

type organizationResource struct {
	client   *ptclient.Client
	provider *client.Provider
}

func (r *organizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
		if client, err := provider.PartnerClient(); err == nil {
			r.client = client
		} else {
			resp.Diagnostics.AddError("Error configuring organization resource", err.Error())
		}
	}
}

func (r *organizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "polytomic_organization"
}

func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data organizationResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Organization.Create(ctx,
		&polytomic.CreateOrganizationRequestSchema{
			Name:      data.Name.ValueString(),
			SsoDomain: data.SSODomain.ValueStringPointer(),
			SsoOrgId:  data.SSOOrgId.ValueStringPointer(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating organization: %s", err))
		return
	}
	data.Id = types.StringPointerValue(created.Data.Id)
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
	organization, err := r.client.Organization.Get(ctx, data.Id.ValueString())
	if err != nil {
		pErr := &ptcore.APIError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading organization: %s", err))
		return
	}

	data.Id = types.StringPointerValue(organization.Data.Id)
	data.Name = types.StringPointerValue(organization.Data.Name)

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

	updated, err := r.client.Organization.Update(ctx, data.Id.ValueString(),
		&polytomic.UpdateOrganizationRequestSchema{
			Name:      data.Name.ValueString(),
			SsoDomain: data.SSODomain.ValueStringPointer(),
			SsoOrgId:  data.SSOOrgId.ValueStringPointer(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating organization: %s", err))
		return
	}

	data.Name = types.StringPointerValue(updated.Data.Name)
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

	err := r.client.Organization.Remove(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting organization", err.Error())
		return
	}
}

func (r *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
