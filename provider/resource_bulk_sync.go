package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/polytomic/polytomic-go"
	"github.com/polytomic/polytomic-go/bulksync"
	ptclient "github.com/polytomic/polytomic-go/client"
	ptcore "github.com/polytomic/polytomic-go/core"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &bulkSyncResource{}
var _ resource.ResourceWithImportState = &bulkSyncResource{}

func (r *bulkSyncResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Bulk Syncs: Bulk Sync",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"dest_connection_id": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"source_connection_id": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"discover": schema.BoolAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"schemas": schema.SetAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"policies": schema.SetAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"schedule": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"frequency": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
					},
					"day_of_week": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"hour": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"minute": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"month": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"day_of_month": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Required: true,
			},
			"dest_configuration": schema.MapAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
				// PlanModifiers: []planmodifier.Map{
				// 	advancedKeyModifier{},
				// 	mapplanmodifier.UseStateForUnknown(),
				// },
			},
			"source_configuration": schema.MapAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
			},
		},
	}
}

func (r *bulkSyncResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ptclient.Client)

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
	Organization             types.String `tfsdk:"organization"`
	Id                       types.String `tfsdk:"id"`
	DestConnectionID         types.String `tfsdk:"dest_connection_id"`
	SourceConnectionID       types.String `tfsdk:"source_connection_id"`
	Mode                     types.String `tfsdk:"mode"`
	Discover                 types.Bool   `tfsdk:"discover"`
	Active                   types.Bool   `tfsdk:"active"`
	Schemas                  types.Set    `tfsdk:"schemas"`
	Policies                 types.Set    `tfsdk:"policies"`
	Schedule                 types.Object `tfsdk:"schedule"`
	DestinationConfiguration types.Map    `tfsdk:"dest_configuration"`
	SourceConfiguration      types.Map    `tfsdk:"source_configuration"`
}

type bulkSyncResource struct {
	client *ptclient.Client
}

type BulkSchedule struct {
	DayOfMonth *string `json:"day_of_month" url:"day_of_month,omitempty" tfsdk:"day_of_month"`
	DayOfWeek  *string `json:"day_of_week" url:"day_of_week,omitempty" tfsdk:"day_of_week"`
	Frequency  string  `json:"frequency" url:"frequency,omitempty" tfsdk:"frequency"`
	Hour       *string `json:"hour" url:"hour,omitempty" tfsdk:"hour"`
	Minute     *string `json:"minute" url:"minute,omitempty" tfsdk:"minute"`
	Month      *string `json:"month" url:"month,omitempty" tfsdk:"month"`
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

	var policies []string
	diags = data.Policies.ElementsAs(ctx, &policies, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var schedule BulkSchedule
	diags = data.Schedule.As(ctx, &schedule, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	sche := &polytomic.BulkSchedule{
		DayOfMonth: schedule.DayOfMonth,
		DayOfWeek:  schedule.DayOfWeek,
		Frequency:  polytomic.ScheduleFrequency(schedule.Frequency),
		Hour:       schedule.Hour,
		Minute:     schedule.Minute,
		Month:      schedule.Month,
	}

	destConfigRaw := make(map[string]string)
	diags = data.DestinationConfiguration.ElementsAs(ctx, &destConfigRaw, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	destConf := make(map[string]interface{})
	for k, v := range destConfigRaw {
		if k == "advanced" {
			var advanced map[string]interface{}
			err := json.Unmarshal([]byte(v), &advanced)
			if err != nil {
				resp.Diagnostics.AddError("Error unmarshalling advanced", err.Error())
				return
			}
			destConf[k] = advanced
		} else {
			destConf[k] = v
		}
	}

	sourceConfigRaw := make(map[string]string)
	diags = data.SourceConfiguration.ElementsAs(ctx, &sourceConfigRaw, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	sourceConf := make(map[string]interface{})
	for k, v := range sourceConfigRaw {
		if k == "advanced" {
			var advanced map[string]interface{}
			err := json.Unmarshal([]byte(v), &advanced)
			if err != nil {
				resp.Diagnostics.AddError("Error unmarshalling advanced", err.Error())
				return
			}
			sourceConf[k] = v
		} else {
			sourceConf[k] = v
		}
	}

	schemaItems := make([]*polytomic.V2CreateBulkSyncRequestSchemasItem, len(schemas))
	for i, s := range schemas {
		schemaItems[i] = &polytomic.V2CreateBulkSyncRequestSchemasItem{
			String: s,
		}
	}
	created, err := r.client.BulkSync.Create(ctx,
		&polytomic.CreateBulkSyncRequest{
			OrganizationId:           data.Organization.ValueStringPointer(),
			Name:                     data.Name.ValueString(),
			DestinationConnectionId:  data.DestConnectionID.ValueString(),
			SourceConnectionId:       data.SourceConnectionID.ValueString(),
			Mode:                     data.Mode.ValueString(),
			Discover:                 data.Discover.ValueBoolPointer(),
			Active:                   data.Active.ValueBoolPointer(),
			Schemas:                  schemaItems,
			Policies:                 policies,
			Schedule:                 sche,
			DestinationConfiguration: destConf,
			SourceConfiguration:      sourceConf,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating bulk sync: %s", err))
		return
	}

	sch, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"frequency":    types.StringType,
		"day_of_week":  types.StringType,
		"hour":         types.StringType,
		"minute":       types.StringType,
		"month":        types.StringType,
		"day_of_month": types.StringType,
	}, BulkSchedule{
		DayOfMonth: created.Data.Schedule.DayOfMonth,
		DayOfWeek:  created.Data.Schedule.DayOfWeek,
		Frequency:  string(created.Data.Schedule.Frequency),
		Hour:       created.Data.Schedule.Hour,
		Minute:     created.Data.Schedule.Minute,
		Month:      created.Data.Schedule.Month,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	schemaVal, diags := types.SetValueFrom(ctx, types.StringType, schemas)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	sourceConfRaw := make(map[string]string)
	for k, v := range created.Data.SourceConfiguration {
		if k == "advanced" {
			advanced, err := json.Marshal(v)
			if err != nil {
				resp.Diagnostics.AddError("Error marshalling advanced", err.Error())
				return
			}
			sourceConfRaw[k] = string(advanced)
		} else {
			sourceConfRaw[k] = stringy(v)
		}
	}

	sourceConfVal, diags := types.MapValueFrom(ctx, types.StringType, sourceConfRaw)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	destConfFinal := make(map[string]string)
	for k, v := range destConf {
		if k == "advanced" {
			advanced, err := json.Marshal(v)
			if err != nil {
				resp.Diagnostics.AddError("Error marshalling advanced", err.Error())
				return
			}
			destConfFinal[k] = string(advanced)
		} else {
			destConfFinal[k] = stringy(v)
		}
	}
	destConfVal, diags := types.MapValueFrom(ctx, types.StringType, destConfFinal)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringPointerValue(created.Data.Id)
	data.Organization = types.StringPointerValue(created.Data.OrganizationId)
	data.Name = types.StringPointerValue(created.Data.Name)
	data.DestConnectionID = types.StringPointerValue(created.Data.DestinationConnectionId)
	data.SourceConnectionID = types.StringPointerValue(created.Data.SourceConnectionId)
	data.Mode = types.StringPointerValue(created.Data.Mode)
	data.Discover = types.BoolPointerValue(created.Data.Discover)
	data.Active = types.BoolPointerValue(created.Data.Active)
	data.Schedule = sch
	data.Schemas = schemaVal
	data.SourceConfiguration = sourceConfVal
	data.DestinationConfiguration = destConfVal

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

	bulkSync, err := r.client.BulkSync.Get(ctx, data.Id.ValueString(), &polytomic.BulkSyncGetRequest{})
	if err != nil {
		pErr := &ptcore.APIError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading bulk sync: %s", err))
		return
	}

	schedule, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"frequency":    types.StringType,
		"day_of_week":  types.StringType,
		"hour":         types.StringType,
		"minute":       types.StringType,
		"month":        types.StringType,
		"day_of_month": types.StringType,
	}, BulkSchedule{
		DayOfMonth: bulkSync.Data.Schedule.DayOfMonth,
		DayOfWeek:  bulkSync.Data.Schedule.DayOfWeek,
		Frequency:  string(bulkSync.Data.Schedule.Frequency),
		Hour:       bulkSync.Data.Schedule.Hour,
		Minute:     bulkSync.Data.Schedule.Minute,
		Month:      bulkSync.Data.Schedule.Month,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Get schemas
	var schemas []*string
	schemasRes, err := r.client.BulkSync.Schemas.List(ctx, data.Id.ValueString(), &bulksync.SchemasListRequest{})
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading bulk sync schemas: %s", err))
		return
	}
	for _, schema := range schemasRes.Data {
		if pointer.GetBool(schema.Enabled) {
			schemas = append(schemas, schema.Id)
		}
	}

	schemaValue, diags := types.SetValueFrom(ctx, types.StringType, schemas)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	sourceConfRaw := make(map[string]string)
	for k, v := range bulkSync.Data.SourceConfiguration {
		if k == "advanced" {
			advanced, err := json.Marshal(v)
			if err != nil {
				resp.Diagnostics.AddError("Error marshalling advanced", err.Error())
				return
			}
			sourceConfRaw[k] = string(advanced)
		} else {
			sourceConfRaw[k] = stringy(v)
		}
	}

	sourceConfVal, diags := types.MapValueFrom(ctx, types.StringType, sourceConfRaw)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	destConfRaw := make(map[string]string)
	for k, v := range bulkSync.Data.DestinationConfiguration {
		if k == "advanced" {
			advanced, err := json.Marshal(v)
			if err != nil {
				resp.Diagnostics.AddError("Error marshalling advanced", err.Error())
				return
			}
			destConfRaw[k] = string(advanced)
		} else {
			destConfRaw[k] = stringy(v)
		}
	}
	destConfVal, diags := types.MapValueFrom(ctx, types.StringType, destConfRaw)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringPointerValue(bulkSync.Data.Id)
	data.Organization = types.StringPointerValue(bulkSync.Data.OrganizationId)
	data.Name = types.StringPointerValue(bulkSync.Data.Name)
	data.DestConnectionID = types.StringPointerValue(bulkSync.Data.DestinationConnectionId)
	data.SourceConnectionID = types.StringPointerValue(bulkSync.Data.SourceConnectionId)
	data.Mode = types.StringPointerValue(bulkSync.Data.Mode)
	data.Discover = types.BoolPointerValue(bulkSync.Data.Discover)
	data.Active = types.BoolPointerValue(bulkSync.Data.Active)
	data.Schedule = schedule
	data.Schemas = schemaValue
	data.SourceConfiguration = sourceConfVal
	data.DestinationConfiguration = destConfVal

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

	var schedule BulkSchedule
	diags = data.Schedule.As(ctx, &schedule, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	sche := &polytomic.BulkSchedule{
		DayOfMonth: schedule.DayOfMonth,
		DayOfWeek:  schedule.DayOfWeek,
		Frequency:  polytomic.ScheduleFrequency(schedule.Frequency),
		Hour:       schedule.Hour,
		Minute:     schedule.Minute,
		Month:      schedule.Month,
	}

	var schemas []string
	diags = data.Schemas.ElementsAs(ctx, &schemas, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	sort.Strings(schemas)

	var policies []string
	diags = data.Policies.ElementsAs(ctx, &policies, false)
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
	destConf := make(map[string]interface{})
	for k, v := range destConfigRaw {
		if k == "advanced" {
			var advanced map[string]interface{}
			err := json.Unmarshal([]byte(v), &advanced)
			if err != nil {
				resp.Diagnostics.AddError("Error unmarshalling advanced", err.Error())
				return
			}
			destConf[k] = advanced
		} else {
			destConf[k] = v
		}
	}

	sourceConfigRaw := make(map[string]string)
	diags = data.SourceConfiguration.ElementsAs(ctx, &sourceConfigRaw, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	sourceConf := make(map[string]interface{})
	for k, v := range sourceConfigRaw {
		if k == "advanced" {
			var advanced map[string]interface{}
			err := json.Unmarshal([]byte(v), &advanced)
			if err != nil {
				resp.Diagnostics.AddError("Error unmarshalling advanced", err.Error())
				return
			}
			sourceConf[k] = v
		} else {
			sourceConf[k] = v
		}
	}

	schemaItems := make([]*polytomic.V2UpdateBulkSyncRequestSchemasItem, len(schemas))
	for i, s := range schemas {
		schemaItems[i] = &polytomic.V2UpdateBulkSyncRequestSchemasItem{
			String: s,
		}
	}
	updated, err := r.client.BulkSync.Update(ctx,
		data.Id.ValueString(),
		&polytomic.UpdateBulkSyncRequest{
			OrganizationId:           data.Organization.ValueStringPointer(),
			Name:                     data.Name.ValueString(),
			DestinationConnectionId:  data.DestConnectionID.ValueString(),
			SourceConnectionId:       data.SourceConnectionID.ValueString(),
			Mode:                     data.Mode.ValueString(),
			Discover:                 data.Discover.ValueBoolPointer(),
			Active:                   data.Active.ValueBoolPointer(),
			Schemas:                  schemaItems,
			Policies:                 policies,
			Schedule:                 sche,
			DestinationConfiguration: destConf,
			SourceConfiguration:      sourceConf,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating organization: %s", err))
		return
	}

	// Get schemas
	var respSchemas []*string
	schemasRes, err := r.client.BulkSync.Schemas.List(ctx, data.Id.ValueString(), &bulksync.SchemasListRequest{Filters: map[string]*string{
		"enabled": pointer.ToString("true"),
	}})
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading bulk sync schemas: %s", err))
		return
	}
	for _, schema := range schemasRes.Data {
		if pointer.GetBool(schema.Enabled) {
			respSchemas = append(respSchemas, schema.Id)
		}
	}

	sch, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"frequency":    types.StringType,
		"day_of_week":  types.StringType,
		"hour":         types.StringType,
		"minute":       types.StringType,
		"month":        types.StringType,
		"day_of_month": types.StringType,
	}, BulkSchedule{
		DayOfMonth: updated.Data.Schedule.DayOfMonth,
		DayOfWeek:  updated.Data.Schedule.DayOfWeek,
		Frequency:  string(updated.Data.Schedule.Frequency),
		Hour:       updated.Data.Schedule.Hour,
		Minute:     updated.Data.Schedule.Minute,
		Month:      updated.Data.Schedule.Month,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	schemaValue, diags := types.SetValueFrom(ctx, types.StringType, respSchemas)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	sourceConfRaw := make(map[string]string)
	for k, v := range updated.Data.SourceConfiguration {
		if k == "advanced" {
			advanced, err := json.Marshal(v)
			if err != nil {
				resp.Diagnostics.AddError("Error marshalling advanced", err.Error())
				return
			}
			sourceConfRaw[k] = string(advanced)
		} else {
			sourceConfRaw[k] = stringy(v)
		}
	}

	sourceConfVal, diags := types.MapValueFrom(ctx, types.StringType, sourceConfRaw)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	destConfFinal := make(map[string]string)
	for k, v := range destConf {
		if k == "advanced" {
			advanced, err := json.Marshal(v)
			if err != nil {
				resp.Diagnostics.AddError("Error marshalling advanced", err.Error())
				return
			}
			destConfFinal[k] = string(advanced)
		} else {
			destConfFinal[k] = stringy(v)
		}
	}
	destConfVal, diags := types.MapValueFrom(ctx, types.StringType, destConfFinal)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringPointerValue(updated.Data.Id)
	data.Organization = types.StringPointerValue(updated.Data.OrganizationId)
	data.Name = types.StringPointerValue(updated.Data.Name)
	data.DestConnectionID = types.StringPointerValue(updated.Data.DestinationConnectionId)
	data.SourceConnectionID = types.StringPointerValue(updated.Data.SourceConnectionId)
	data.Mode = types.StringPointerValue(updated.Data.Mode)
	data.Discover = types.BoolPointerValue(updated.Data.Discover)
	data.Active = types.BoolPointerValue(updated.Data.Active)
	data.Schedule = sch
	data.Schemas = schemaValue
	data.SourceConfiguration = sourceConfVal
	data.DestinationConfiguration = destConfVal

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

func (r *bulkSyncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bulkSyncResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.BulkSync.Remove(ctx, data.Id.ValueString(), &polytomic.BulkSyncRemoveRequest{})
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting organization: %s", err))
		return
	}
}

func (r *bulkSyncResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
