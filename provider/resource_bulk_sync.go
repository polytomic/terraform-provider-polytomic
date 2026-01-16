package provider

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/polytomic/polytomic-go"
	"github.com/polytomic/polytomic-go/bulksync"
	ptcore "github.com/polytomic/polytomic-go/core"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &bulkSyncResource{}
var _ resource.ResourceWithImportState = &bulkSyncResource{}

// NewBulkSyncResourceForSchemaIntrospection returns a bulk sync resource instance
// for schema introspection. This is used by the importer to validate field mappings.
func NewBulkSyncResourceForSchemaIntrospection() resource.Resource {
	return &bulkSyncResource{}
}

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
			"active": schema.BoolAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"source": schema.SingleNestedAttribute{
				Required:   true,
				Attributes: bulkSyncConnection{}.SchemaAttributes(),
			},
			"destination": schema.SingleNestedAttribute{
				Required:   true,
				Attributes: bulkSyncConnection{}.SchemaAttributes(),
			},
			"automatically_add_new_fields": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Default:  stringdefault.StaticString(string(polytomic.BulkDiscoverNone)),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // Ensures state carries the default if unset
				},
			},
			"automatically_add_new_objects": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Default:  stringdefault.StaticString(string(polytomic.BulkDiscoverNone)),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(), // Ensures state carries the default if unset
				},
			},
			"disable_record_timestamps": schema.BoolAttribute{
				Optional: true,
				Computed: true,
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
			"data_cutoff_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Optional:   true,
			},
			"schemas": schema.SetNestedAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: bulkSyncSchema{}.SchemaAttributes(),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"policies": schema.SetAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"concurrency_limit": schema.Int64Attribute{
				MarkdownDescription: "Per-sync concurrency limit override",
				Optional:            true,
				Computed:            true,
			},
			"resync_concurrency_limit": schema.Int64Attribute{
				MarkdownDescription: "Per-sync resync concurrency limit override",
				Optional:            true,
				Computed:            true,
			},
			"normalize_names": schema.StringAttribute{
				MarkdownDescription: "Name normalization settings",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

type bulkSyncConnection struct {
	ConnectionID  types.String         `tfsdk:"connection_id"`
	Configuration jsontypes.Normalized `tfsdk:"configuration"`
}

func (bulkSyncConnection) SchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"connection_id": schema.StringAttribute{
			Required: true,
		},
		"configuration": schema.StringAttribute{
			MarkdownDescription: "Integration-specific configuration for the connection. Documentation for settings is available in the [Polytomic API documentation](https://apidocs.polytomic.com/2024-02-08/guides/configuring-your-connections/overview)",
			CustomType:          jsontypes.NormalizedType{},
			Required:            true,
		},
	}
}

func (bulkSyncConnection) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"connection_id": types.StringType,
		"configuration": jsontypes.NormalizedType{},
	}
}

type bulkSyncSchemaField struct {
	Id        types.String `tfsdk:"id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Obfuscate types.Bool   `tfsdk:"obfuscate"`
}

func (bulkSyncSchemaField) SchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"enabled": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"obfuscate": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
	}
}

func (bulkSyncSchemaField) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":        types.StringType,
		"enabled":   types.BoolType,
		"obfuscate": types.BoolType,
	}
}

type bulkSyncSchema struct {
	Id                  types.String      `tfsdk:"id"`
	Enabled             types.Bool        `tfsdk:"enabled"`
	PartitionKey        types.String      `tfsdk:"partition_key"`
	TrackingField       types.String      `tfsdk:"tracking_field"`
	OutputName          types.String      `tfsdk:"output_name"`
	Fields              types.Set         `tfsdk:"fields"`
	Filters             types.Set         `tfsdk:"filters"`
	DataCutoffTimestamp timetypes.RFC3339 `tfsdk:"data_cutoff_timestamp"`
	DisableDataCutoff   types.Bool        `tfsdk:"disable_data_cutoff"`
}

func (bulkSyncSchema) SchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"enabled": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"data_cutoff_timestamp": schema.StringAttribute{
			CustomType: timetypes.RFC3339Type{},
			Optional:   true,
		},
		"disable_data_cutoff": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"partition_key": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"tracking_field": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"output_name": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"fields": schema.SetNestedAttribute{
			Optional: true,
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: bulkSyncSchemaField{}.SchemaAttributes(),
			},
		},
		"filters": schema.SetNestedAttribute{
			Optional: true,
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"field_id": schema.StringAttribute{
						Optional: true,
					},
					"function": schema.StringAttribute{
						Required: true,
					},
					"value": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func (bulkSyncSchema) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":      types.StringType,
		"enabled": types.BoolType,
		"fields": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: bulkSyncSchemaField{}.AttrTypes(),
			},
		},
		"filters": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"field_id": types.StringType,
					"function": types.StringType,
					"value":    types.StringType,
				},
			},
		},
		"data_cutoff_timestamp": timetypes.RFC3339Type{},
		"disable_data_cutoff":   types.BoolType,
		"partition_key":         types.StringType,
		"tracking_field":        types.StringType,
		"output_name":           types.StringType,
	}
}

type bulkSyncResourceData struct {
	Id                         types.String      `tfsdk:"id"`
	Organization               types.String      `tfsdk:"organization"`
	Name                       types.String      `tfsdk:"name"`
	Active                     types.Bool        `tfsdk:"active"`
	Mode                       types.String      `tfsdk:"mode"`
	Source                     types.Object      `tfsdk:"source"`
	Destination                types.Object      `tfsdk:"destination"`
	AutomaticallyAddNewFields  types.String      `tfsdk:"automatically_add_new_fields"`
	AutomaticallyAddNewObjects types.String      `tfsdk:"automatically_add_new_objects"`
	DisableRecordTimestamps    types.Bool        `tfsdk:"disable_record_timestamps"`
	Schedule                   types.Object      `tfsdk:"schedule"`
	Schemas                    types.Set         `tfsdk:"schemas"`
	Policies                   types.Set         `tfsdk:"policies"`
	DataCutoffTimestamp        timetypes.RFC3339 `tfsdk:"data_cutoff_timestamp"`
	ConcurrencyLimit           types.Int64       `tfsdk:"concurrency_limit"`
	ResyncConcurrencyLimit     types.Int64       `tfsdk:"resync_concurrency_limit"`
	NormalizeNames             types.String      `tfsdk:"normalize_names"`
}

type bulkSyncResource struct {
	provider *providerclient.Provider
}

func (r *bulkSyncResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *bulkSyncResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_sync"
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

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var schemaData []bulkSyncSchema
	diags = data.Schemas.ElementsAs(ctx, &schemaData, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	slices.SortFunc(schemaData, func(a, b bulkSyncSchema) int {
		return cmp.Compare(a.Id.String(), b.Id.String())
	})

	schemas := make([]*polytomic.V2CreateBulkSyncRequestSchemasItem, len(schemaData))
	for i, s := range schemaData {
		var cutoff *time.Time
		if !s.DataCutoffTimestamp.IsNull() {
			t, diags := s.DataCutoffTimestamp.ValueRFC3339Time()
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			cutoff = &t
		}

		// fields
		var fieldData []bulkSyncSchemaField
		if !s.Fields.IsUnknown() {
			diags = s.Fields.ElementsAs(ctx, &fieldData, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			slices.SortFunc(fieldData, func(a, b bulkSyncSchemaField) int {
				return cmp.Compare(a.Id.String(), b.Id.String())
			})
		}
		fieldConfs := make([]*polytomic.V2SchemaConfigurationFieldsItem, len(fieldData))
		for i, f := range fieldData {
			fieldConfs[i] = &polytomic.V2SchemaConfigurationFieldsItem{
				FieldConfiguration: &polytomic.FieldConfiguration{
					Id:        f.Id.ValueStringPointer(),
					Enabled:   f.Enabled.ValueBoolPointer(),
					Obfuscate: f.Obfuscate.ValueBoolPointer(),
				},
			}
		}

		// filters
		var filters []*polytomic.BulkFilter
		if !s.Filters.IsUnknown() {
			diags = s.Filters.ElementsAs(ctx, &filters, false)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
		}

		schemas[i] = &polytomic.V2CreateBulkSyncRequestSchemasItem{
			SchemaConfiguration: &polytomic.SchemaConfiguration{
				Id:                  s.Id.ValueStringPointer(),
				Enabled:             s.Enabled.ValueBoolPointer(),
				PartitionKey:        s.PartitionKey.ValueStringPointer(),
				DataCutoffTimestamp: cutoff,
				DisableDataCutoff:   s.DisableDataCutoff.ValueBoolPointer(),
				Fields:              fieldConfs,
				Filters:             filters,
			},
		}
	}

	// policies
	var policies []string
	if !(data.Policies.IsUnknown() || data.Policies.IsNull()) {
		diags = data.Policies.ElementsAs(ctx, &policies, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	// schedule
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

	// destination connection
	destination := bulkSyncConnection{}
	diags = data.Destination.As(ctx, &destination, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	var destConf map[string]interface{}
	if !destination.Configuration.IsUnknown() {
		dc := make(map[string]interface{})
		diags = destination.Configuration.Unmarshal(&dc)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		destConf = dc
	}

	// source connection
	source := bulkSyncConnection{}
	diags = data.Source.As(ctx, &source, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	var sourceConf map[string]interface{}
	if !source.Configuration.IsUnknown() {
		sc := make(map[string]interface{})
		diags = source.Configuration.Unmarshal(&sc)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		sourceConf = sc
	}
	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	// Convert int64 concurrency limits to int
	var concurrencyLimit *int
	if !data.ConcurrencyLimit.IsNull() {
		val := int(data.ConcurrencyLimit.ValueInt64())
		concurrencyLimit = &val
	}
	var resyncConcurrencyLimit *int
	if !data.ResyncConcurrencyLimit.IsNull() {
		val := int(data.ResyncConcurrencyLimit.ValueInt64())
		resyncConcurrencyLimit = &val
	}

	created, err := client.BulkSync.Create(ctx,
		&polytomic.CreateBulkSyncRequest{
			OrganizationId:             data.Organization.ValueStringPointer(),
			Name:                       data.Name.ValueString(),
			DestinationConnectionId:    destination.ConnectionID.ValueString(),
			SourceConnectionId:         source.ConnectionID.ValueString(),
			Mode:                       pointer.To(polytomic.BulkSyncMode(data.Mode.ValueString())),
			Active:                     data.Active.ValueBoolPointer(),
			AutomaticallyAddNewFields:  pointer.To(polytomic.BulkDiscover(data.AutomaticallyAddNewFields.ValueString())),
			AutomaticallyAddNewObjects: pointer.To(polytomic.BulkDiscover(data.AutomaticallyAddNewObjects.ValueString())),
			Schemas:                    schemas,
			Policies:                   policies,
			Schedule:                   sche,
			DestinationConfiguration:   destConf,
			SourceConfiguration:        sourceConf,
			ConcurrencyLimit:           concurrencyLimit,
			ResyncConcurrencyLimit:     resyncConcurrencyLimit,
			NormalizeNames:             pointer.To(polytomic.BulkNormalizeNames(data.NormalizeNames.ValueString())),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error creating bulk sync: %s", err))
		return
	}
	createdSchemas, err := client.BulkSync.Schemas.List(ctx, pointer.Get(created.Data.Id), &bulksync.SchemasListRequest{})
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error reading bulk sync schemas: %s", err))
		return
	}
	data, diags = bulkSyncDataFromResponse(ctx, created.Data, createdSchemas.Data)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *bulkSyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bulkSyncResourceData

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
	bulkSync, err := client.BulkSync.Get(ctx, data.Id.ValueString(), &polytomic.BulkSyncGetRequest{})
	if err != nil {
		pErr := &ptcore.APIError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error reading bulk sync: %s", err))
		return
	}
	schemas, err := client.BulkSync.Schemas.List(ctx, data.Id.ValueString(), &bulksync.SchemasListRequest{})
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error reading bulk sync schemas: %s", err))
		return
	}
	data, diags = bulkSyncDataFromResponse(ctx, bulkSync.Data, schemas.Data)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
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

	var schemaData []bulkSyncSchema
	diags = data.Schemas.ElementsAs(ctx, &schemaData, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	slices.SortFunc(schemaData, func(a, b bulkSyncSchema) int {
		return cmp.Compare(a.Id.String(), b.Id.String())
	})

	schemas := make([]*polytomic.V2UpdateBulkSyncRequestSchemasItem, len(schemaData))
	for i, s := range schemaData {
		var cutoff *time.Time
		if !s.DataCutoffTimestamp.IsNull() {
			t, diags := s.DataCutoffTimestamp.ValueRFC3339Time()
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			cutoff = &t
		}

		// fields
		var fields []polytomic.FieldConfiguration
		diags = s.Fields.ElementsAs(ctx, &fields, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		fieldConfs := make([]*polytomic.V2SchemaConfigurationFieldsItem, len(fields))
		for i, f := range fields {
			fieldConfs[i] = &polytomic.V2SchemaConfigurationFieldsItem{
				FieldConfiguration: &f,
			}
		}

		// filters
		var filters []*polytomic.BulkFilter
		diags = s.Filters.ElementsAs(ctx, &filters, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		schemas[i] = &polytomic.V2UpdateBulkSyncRequestSchemasItem{
			SchemaConfiguration: &polytomic.SchemaConfiguration{
				Id:                  s.Id.ValueStringPointer(),
				Enabled:             s.Enabled.ValueBoolPointer(),
				PartitionKey:        s.PartitionKey.ValueStringPointer(),
				DataCutoffTimestamp: cutoff,
				DisableDataCutoff:   s.DisableDataCutoff.ValueBoolPointer(),
				Fields:              fieldConfs,
				Filters:             filters,
			},
		}
	}

	// policies
	var policies []string
	diags = data.Policies.ElementsAs(ctx, &policies, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// schedule
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

	// destination connection
	destination := bulkSyncConnection{}
	diags = data.Destination.As(ctx, &destination, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	destConf := make(map[string]interface{})
	diags = destination.Configuration.Unmarshal(&destConf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// source connection
	source := bulkSyncConnection{}
	diags = data.Source.As(ctx, &source, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	sourceConf := make(map[string]interface{})
	diags = source.Configuration.Unmarshal(&sourceConf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	// Convert int64 concurrency limits to int
	var concurrencyLimit *int
	if !data.ConcurrencyLimit.IsNull() {
		val := int(data.ConcurrencyLimit.ValueInt64())
		concurrencyLimit = &val
	}
	var resyncConcurrencyLimit *int
	if !data.ResyncConcurrencyLimit.IsNull() {
		val := int(data.ResyncConcurrencyLimit.ValueInt64())
		resyncConcurrencyLimit = &val
	}

	updated, err := client.BulkSync.Update(ctx,
		data.Id.ValueString(),
		&polytomic.UpdateBulkSyncRequest{
			OrganizationId:             data.Organization.ValueStringPointer(),
			Name:                       data.Name.ValueString(),
			DestinationConnectionId:    destination.ConnectionID.ValueString(),
			SourceConnectionId:         source.ConnectionID.ValueString(),
			Mode:                       pointer.To(polytomic.BulkSyncMode(data.Mode.ValueString())),
			Active:                     data.Active.ValueBoolPointer(),
			AutomaticallyAddNewFields:  pointer.To(polytomic.BulkDiscover(data.AutomaticallyAddNewFields.ValueString())),
			AutomaticallyAddNewObjects: pointer.To(polytomic.BulkDiscover(data.AutomaticallyAddNewObjects.ValueString())),
			Schemas:                    schemas,
			Policies:                   policies,
			Schedule:                   sche,
			DestinationConfiguration:   destConf,
			SourceConfiguration:        sourceConf,
			ConcurrencyLimit:           concurrencyLimit,
			ResyncConcurrencyLimit:     resyncConcurrencyLimit,
			NormalizeNames:             pointer.To(polytomic.BulkNormalizeNames(data.NormalizeNames.ValueString())),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error creating bulk sync: %s", err))
		return
	}
	updatedSchemas, err := client.BulkSync.Schemas.List(ctx, data.Id.ValueString(), &bulksync.SchemasListRequest{})
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error reading bulk sync schemas: %s", err))
		return
	}

	data, diags = bulkSyncDataFromResponse(ctx, updated.Data, updatedSchemas.Data)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

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

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	err = client.BulkSync.Remove(ctx, data.Id.ValueString(), &polytomic.BulkSyncRemoveRequest{})
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error deleting organization: %s", err))
		return
	}
}

func (r *bulkSyncResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// bulkSyncDataFromResponse returns the Terraform data for the response from the
// Polytomic API.
func bulkSyncDataFromResponse(ctx context.Context, response *polytomic.BulkSyncResponse, schemas []*polytomic.BulkSchema) (bulkSyncResourceData, diag.Diagnostics) {
	var data bulkSyncResourceData
	// schedule result
	sch, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"frequency":    types.StringType,
		"day_of_week":  types.StringType,
		"hour":         types.StringType,
		"minute":       types.StringType,
		"month":        types.StringType,
		"day_of_month": types.StringType,
	}, BulkSchedule{
		DayOfMonth: response.Schedule.DayOfMonth,
		DayOfWeek:  response.Schedule.DayOfWeek,
		Frequency:  string(response.Schedule.Frequency),
		Hour:       response.Schedule.Hour,
		Minute:     response.Schedule.Minute,
		Month:      response.Schedule.Month,
	})
	if diags.HasError() {
		return data, diags
	}

	// schemas result
	schemaVal, diags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: bulkSyncSchema{}.AttrTypes()}, schemas)
	if diags.HasError() {
		return data, diags
	}

	// source connection result
	sourceConfVal, err := json.Marshal(response.SourceConfiguration)
	if err != nil {
		diags.AddError("Error reading source configuration", err.Error())
		return data, diags
	}
	sourceVal, diags := types.ObjectValueFrom(ctx, bulkSyncConnection{}.AttrTypes(),
		bulkSyncConnection{
			ConnectionID:  types.StringPointerValue(response.SourceConnectionId),
			Configuration: jsontypes.NewNormalizedValue(string(sourceConfVal)),
		},
	)
	if diags.HasError() {
		return data, diags
	}

	// destination connection result
	destConfVal, err := json.Marshal(response.DestinationConfiguration)
	if err != nil {
		diags.AddError("Error reading source configuration", err.Error())
		return data, diags
	}
	destVal, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"connection_id": types.StringType,
		"configuration": jsontypes.NormalizedType{},
	},
		bulkSyncConnection{
			ConnectionID:  types.StringPointerValue(response.DestinationConnectionId),
			Configuration: jsontypes.NewNormalizedValue(string(destConfVal)),
		},
	)
	if diags.HasError() {
		return data, diags
	}

	data.Id = types.StringPointerValue(response.Id)
	data.Organization = types.StringPointerValue(response.OrganizationId)
	data.Name = types.StringPointerValue(response.Name)
	data.Mode = types.StringValue(string(pointer.Get(response.Mode)))
	data.Active = types.BoolPointerValue(response.Active)
	data.AutomaticallyAddNewFields = types.StringValue(string(pointer.Get((response.AutomaticallyAddNewFields))))
	data.AutomaticallyAddNewObjects = types.StringValue(string(pointer.Get((response.AutomaticallyAddNewObjects))))
	data.Destination = destVal
	data.Source = sourceVal
	data.Schedule = sch
	data.Schemas = schemaVal
	data.Policies, _ = types.SetValueFrom(ctx, types.StringType, response.Policies)

	if response.ConcurrencyLimit != nil {
		data.ConcurrencyLimit = types.Int64Value(int64(*response.ConcurrencyLimit))
	} else {
		data.ConcurrencyLimit = types.Int64Null()
	}
	if response.ResyncConcurrencyLimit != nil {
		data.ResyncConcurrencyLimit = types.Int64Value(int64(*response.ResyncConcurrencyLimit))
	} else {
		data.ResyncConcurrencyLimit = types.Int64Null()
	}
	if response.NormalizeNames != nil {
		data.NormalizeNames = types.StringValue(string(*response.NormalizeNames))
	} else {
		data.NormalizeNames = types.StringNull()
	}

	return data, diags
}
