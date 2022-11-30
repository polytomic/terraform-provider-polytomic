package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &bulkSyncResource{}
var _ resource.ResourceWithImportState = &bulkSyncResource{}

func (r *bulkSyncResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
			"name": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"dest_connection_id": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"source_connection_id": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"mode": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"discover": {
				MarkdownDescription: "",
				Type:                types.BoolType,
				Required:            true,
			},
			"active": {
				MarkdownDescription: "",
				Type:                types.BoolType,
				Required:            true,
			},
			"schemas": {
				MarkdownDescription: "",
				Type:                types.ListType{ElemType: types.StringType},
				Optional:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
			"schedule": {
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"frequency": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
					"day_of_week": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
					},
					"hour": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
					},
					"minute": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
					},
					"month": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
					},
					"day_of_month": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
					},
				}),
				Required: true,
			},
			"dest_configuration": {
				MarkdownDescription: `\
				- Snowflake: schema = "" \
				- BigQuery: dataset = "" \
				- S3: format = "csv/json"`,
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Required: true,
			},
			"source_configuration": {
				MarkdownDescription: "",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
		},
	}, nil
}

func (r *bulkSyncResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || !req.State.Raw.IsKnown() {
		return
	}

	config := &bulkSyncResourceData{}
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)

	plan := &bulkSyncResourceData{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keys := []string{"day_of_week", "hour", "minute", "month", "day_of_month"}
	for _, key := range keys {
		if config.Schedule.Attributes()[key].IsNull() {
			plan.Schedule.Attributes()[key] = types.StringValue("")
		}
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)

}

func (r *bulkSyncResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bulkSyncResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_sync"
}

type bulkSyncResourceData struct {
	Name                     types.String `tfsdk:"name"`
	Id                       types.String `tfsdk:"id"`
	DestConnectionID         types.String `tfsdk:"dest_connection_id"`
	SourceConnectionID       types.String `tfsdk:"source_connection_id"`
	Mode                     types.String `tfsdk:"mode"`
	Discover                 types.Bool   `tfsdk:"discover"`
	Active                   types.Bool   `tfsdk:"active"`
	Schemas                  types.List   `tfsdk:"schemas"`
	Schedule                 types.Object `tfsdk:"schedule"`
	DestinationConfiguration types.Map    `tfsdk:"dest_configuration"`
	SourceConfiguration      types.Map    `tfsdk:"source_configuration"`
}

type bulkSyncResource struct {
	client *polytomic.Client
}

func (r *bulkSyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bulkSyncResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var schemas []string
	diags = data.Schemas.ElementsAs(ctx, &schemas, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	sort.Strings(schemas)

	var schedule polytomic.Schedule
	diags = data.Schedule.As(ctx, &schedule, types.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	destConfigRaw := make(map[string]string)
	diags = data.DestinationConfiguration.ElementsAs(ctx, &destConfigRaw, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	destConfig := make(map[string]interface{})
	for k, v := range destConfigRaw {
		destConfig[k] = v
	}

	sourceConfgRaw := make(map[string]string)
	diags = data.SourceConfiguration.ElementsAs(ctx, &sourceConfgRaw, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	sourceConfig := make(map[string]interface{})
	for k, v := range sourceConfgRaw {
		sourceConfig[k] = v
	}

	created, err := r.client.Bulk().CreateBulkSync(ctx,
		polytomic.BulkSyncRequest{
			Name:                     data.Name.ValueString(),
			DestConnectionID:         data.DestConnectionID.ValueString(),
			SourceConnectionID:       data.SourceConnectionID.ValueString(),
			Mode:                     data.Mode.ValueString(),
			Discover:                 data.Discover.ValueBool(),
			Active:                   data.Active.ValueBool(),
			Schemas:                  schemas,
			Schedule:                 schedule,
			DestinationConfiguration: destConfig,
			SourceConfiguration:      sourceConfig,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating bulk sync: %s", err))
		return
	}
	data.Id = types.StringValue(created.ID)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bulkSyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bulkSyncResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	bulkSync, err := r.client.Bulk().GetBulkSync(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading bulk sync: %s", err))
		return
	}

	schedule, diags := types.ObjectValueFrom(ctx,
		map[string]attr.Type{
			"frequency":    types.StringType,
			"day_of_week":  types.StringType,
			"hour":         types.StringType,
			"minute":       types.StringType,
			"month":        types.StringType,
			"day_of_month": types.StringType,
		},
		bulkSync.Schedule)

	data.Id = types.StringValue(bulkSync.ID)
	data.Name = types.StringValue(bulkSync.Name)
	data.DestConnectionID = types.StringValue(bulkSync.DestConnectionID)
	data.SourceConnectionID = types.StringValue(bulkSync.SourceConnectionID)
	data.Mode = types.StringValue(bulkSync.Mode)
	data.Discover = types.BoolValue(bulkSync.Discover)
	data.Active = types.BoolValue(bulkSync.Active)
	data.Schedule = schedule

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

func (r *bulkSyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data bulkSyncResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var schedule polytomic.Schedule
	diags = data.Schedule.As(ctx, &schedule, types.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	updated, err := r.client.Bulk().UpdateBulkSync(ctx,
		data.Id.ValueString(),
		polytomic.BulkSyncRequest{
			Name:               data.Name.ValueString(),
			Active:             data.Active.ValueBool(),
			Discover:           data.Discover.ValueBool(),
			Mode:               data.Mode.ValueString(),
			SourceConnectionID: data.SourceConnectionID.ValueString(),
			DestConnectionID:   data.DestConnectionID.ValueString(),
			Schedule:           schedule,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating organization: %s", err))
		return
	}

	data.Name = types.StringValue(updated.Name)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

	return
}

func (r *bulkSyncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bulkSyncResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Bulk().DeleteBulkSync(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting organization: %s", err))
		return
	}
}

func (r *bulkSyncResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
