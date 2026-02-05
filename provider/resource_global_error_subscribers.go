package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

const globalErrorSubscribersResourceID = "global-error-subscribers"

var _ resource.Resource = &globalErrorSubscribersResource{}
var _ resource.ResourceWithImportState = &globalErrorSubscribersResource{}

func (r *globalErrorSubscribersResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_error_subscribers"
}

func (r *globalErrorSubscribersResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Notifications: Global error subscribers",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization ID (required when using partner or deployment keys)",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"emails": schema.SetAttribute{
				MarkdownDescription: "Email addresses to notify for global errors",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

type globalErrorSubscribersResource struct {
	provider *providerclient.Provider
}

type globalErrorSubscribersResourceModel struct {
	Organization types.String `tfsdk:"organization"`
	Emails       types.Set    `tfsdk:"emails"`
	ID           types.String `tfsdk:"id"`
}

func (r *globalErrorSubscribersResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *globalErrorSubscribersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data globalErrorSubscribersResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	var emails []string
	resp.Diagnostics.Append(data.Emails.ElementsAs(ctx, &emails, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := client.Notifications.SetGlobalErrorSubscribers(
		ctx,
		&polytomic.V4GlobalErrorSubscribersRequest{Emails: emails},
	)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error setting global error subscribers: %s", err))
		return
	}

	emailsToStore := emails
	if response != nil && response.Emails != nil {
		emailsToStore = response.Emails
	}

	emailsSet, diags := types.SetValueFrom(ctx, types.StringType, emailsToStore)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Emails = emailsSet
	data.ID = types.StringValue(globalErrorSubscribersResourceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *globalErrorSubscribersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data globalErrorSubscribersResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	response, err := client.Notifications.GetGlobalErrorSubscribers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error reading global error subscribers: %s", err))
		return
	}
	if response == nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, "Error reading global error subscribers: empty response")
		return
	}

	emailsSet, diags := types.SetValueFrom(ctx, types.StringType, response.Emails)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Emails = emailsSet
	data.ID = types.StringValue(globalErrorSubscribersResourceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *globalErrorSubscribersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data globalErrorSubscribersResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	var emails []string
	resp.Diagnostics.Append(data.Emails.ElementsAs(ctx, &emails, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := client.Notifications.SetGlobalErrorSubscribers(
		ctx,
		&polytomic.V4GlobalErrorSubscribersRequest{Emails: emails},
	)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error updating global error subscribers: %s", err))
		return
	}

	emailsToStore := emails
	if response != nil && response.Emails != nil {
		emailsToStore = response.Emails
	}

	emailsSet, diags := types.SetValueFrom(ctx, types.StringType, emailsToStore)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Emails = emailsSet
	data.ID = types.StringValue(globalErrorSubscribersResourceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *globalErrorSubscribersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data globalErrorSubscribersResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	_, err = client.Notifications.SetGlobalErrorSubscribers(
		ctx,
		&polytomic.V4GlobalErrorSubscribersRequest{Emails: []string{}},
	)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error clearing global error subscribers: %s", err))
	}
}

func (r *globalErrorSubscribersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
