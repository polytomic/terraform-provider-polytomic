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
	ptcore "github.com/polytomic/polytomic-go/core"
	"github.com/polytomic/polytomic-go/permissions"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/providerclient"
)

var _ resource.Resource = &roleResource{}
var _ resource.ResourceWithImportState = &roleResource{}

func (r *roleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Organizations: A role in a Polytomic organization",
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Required: true,
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

func (r roleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

type roleResourceData struct {
	Organization types.String `tfsdk:"organization"`
	Name         types.String `tfsdk:"name"`
	Id           types.String `tfsdk:"id"`
}

type roleResource struct {
	provider *providerclient.Provider
}

func (r *roleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data roleResourceData

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
	role, err := client.Permissions.Roles.Create(
		ctx,
		&permissions.CreateRoleRequest{
			Name:           data.Name.ValueString(),
			OrganizationId: data.Organization.ValueStringPointer(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error creating role: %s", err))
		return
	}
	data.Id = types.StringPointerValue(role.Data.Id)
	data.Name = types.StringPointerValue(role.Data.Name)
	data.Organization = types.StringPointerValue(role.Data.OrganizationId)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data roleResourceData

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
	role, err := client.Permissions.Roles.Get(ctx, data.Id.ValueString())
	if err != nil {
		pErr := &ptcore.APIError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError("Error reading role", err.Error())
		return
	}

	data.Id = types.StringPointerValue(role.Data.Id)
	data.Name = types.StringPointerValue(role.Data.Name)
	data.Organization = types.StringPointerValue(role.Data.OrganizationId)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data roleResourceData

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
	role, err := client.Permissions.Roles.Update(
		ctx,
		data.Id.ValueString(),
		&permissions.UpdateRoleRequest{
			Name:           data.Name.ValueString(),
			OrganizationId: data.Organization.ValueStringPointer(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error updating role: %s", err))
		return
	}

	data.Id = types.StringPointerValue(role.Data.Id)
	data.Name = types.StringPointerValue(role.Data.Name)
	data.Organization = types.StringPointerValue(role.Data.OrganizationId)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data roleResourceData

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
	err = client.Permissions.Roles.Remove(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error deleting role: %s", err))
		return
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
