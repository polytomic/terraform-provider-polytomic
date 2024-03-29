package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/google/uuid"
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

	var policies []string
	diags = data.Policies.ElementsAs(ctx, &policies, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var schedule polytomic.BulkSchedule
	diags = data.Schedule.As(ctx, &schedule, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
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

	created, err := r.client.Bulk().CreateBulkSync(ctx,
		polytomic.BulkSyncRequest{
			OrganizationID:           data.Organization.ValueString(),
			Name:                     data.Name.ValueString(),
			DestConnectionID:         data.DestConnectionID.ValueString(),
			SourceConnectionID:       data.SourceConnectionID.ValueString(),
			Mode:                     data.Mode.ValueString(),
			Discover:                 data.Discover.ValueBool(),
			Active:                   data.Active.ValueBool(),
			Schemas:                  schemas,
			Policies:                 policies,
			Schedule:                 schedule,
			DestinationConfiguration: destConf,
			SourceConfiguration:      sourceConf,
		},
		polytomic.WithIdempotencyKey(uuid.NewString()),
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
	}, created.Schedule)
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
	for k, v := range created.SourceConfiguration {
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

	data.Id = types.StringValue(created.ID)
	data.Organization = types.StringValue(created.OrganizationID)
	data.Name = types.StringValue(created.Name)
	data.DestConnectionID = types.StringValue(created.DestConnectionID)
	data.SourceConnectionID = types.StringValue(created.SourceConnectionID)
	data.Mode = types.StringValue(created.Mode)
	data.Discover = types.BoolValue(created.Discover)
	data.Active = types.BoolValue(created.Active)
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

	bulkSync, err := r.client.Bulk().GetBulkSync(ctx, data.Id.ValueString())
	if err != nil {
		pErr := polytomic.ApiError{}
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
	}, bulkSync.Schedule)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Get schemas
	var schemas []string
	schemasRes, err := r.client.Bulk().GetBulkSyncSchemas(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading bulk sync schemas: %s", err))
		return
	}
	for _, schema := range schemasRes {
		if schema.Enabled {
			schemas = append(schemas, schema.ID)
		}
	}

	schemaValue, diags := types.SetValueFrom(ctx, types.StringType, schemas)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	sourceConfRaw := make(map[string]string)
	for k, v := range bulkSync.SourceConfiguration {
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
	for k, v := range bulkSync.DestinationConfiguration {
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

	data.Id = types.StringValue(bulkSync.ID)
	data.Organization = types.StringValue(bulkSync.OrganizationID)
	data.Name = types.StringValue(bulkSync.Name)
	data.DestConnectionID = types.StringValue(bulkSync.DestConnectionID)
	data.SourceConnectionID = types.StringValue(bulkSync.SourceConnectionID)
	data.Mode = types.StringValue(bulkSync.Mode)
	data.Discover = types.BoolValue(bulkSync.Discover)
	data.Active = types.BoolValue(bulkSync.Active)
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

	var schedule polytomic.BulkSchedule
	diags = data.Schedule.As(ctx, &schedule, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
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

	updated, err := r.client.Bulk().UpdateBulkSync(ctx,
		data.Id.ValueString(),
		polytomic.BulkSyncRequest{
			OrganizationID:           data.Organization.ValueString(),
			Name:                     data.Name.ValueString(),
			DestConnectionID:         data.DestConnectionID.ValueString(),
			SourceConnectionID:       data.SourceConnectionID.ValueString(),
			Mode:                     data.Mode.ValueString(),
			Discover:                 data.Discover.ValueBool(),
			Active:                   data.Active.ValueBool(),
			Schemas:                  schemas,
			Policies:                 policies,
			Schedule:                 schedule,
			DestinationConfiguration: destConf,
			SourceConfiguration:      sourceConf,
		},
		polytomic.WithIdempotencyKey(uuid.NewString()),
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating organization: %s", err))
		return
	}

	// Get schemas
	var respSchemas []string
	schemasRes, err := r.client.Bulk().GetBulkSyncSchemas(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading bulk sync schemas: %s", err))
		return
	}
	for _, schema := range schemasRes {
		if schema.Enabled {
			respSchemas = append(respSchemas, schema.ID)
		}
	}

	sch, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"frequency":    types.StringType,
		"day_of_week":  types.StringType,
		"hour":         types.StringType,
		"minute":       types.StringType,
		"month":        types.StringType,
		"day_of_month": types.StringType,
	}, updated.Schedule)
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
	for k, v := range updated.SourceConfiguration {
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

	data.Id = types.StringValue(updated.ID)
	data.Organization = types.StringValue(updated.OrganizationID)
	data.Name = types.StringValue(updated.Name)
	data.DestConnectionID = types.StringValue(updated.DestConnectionID)
	data.SourceConnectionID = types.StringValue(updated.SourceConnectionID)
	data.Mode = types.StringValue(updated.Mode)
	data.Discover = types.BoolValue(updated.Discover)
	data.Active = types.BoolValue(updated.Active)
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

	err := r.client.Bulk().DeleteBulkSync(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting organization: %s", err))
		return
	}
}

func (r *bulkSyncResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
