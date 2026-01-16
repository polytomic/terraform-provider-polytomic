package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/polytomic/polytomic-go"
	ptcore "github.com/polytomic/polytomic-go/core"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &syncResource{}
var _ resource.ResourceWithImportState = &syncResource{}

// NewSyncResourceForSchemaIntrospection returns a sync resource instance
// for schema introspection. This is used by the importer to validate field mappings.
func NewSyncResourceForSchemaIntrospection() resource.Resource {
	return &syncResource{}
}

func (r *syncResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Syncs: Sync",
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
			"target": schema.SingleNestedAttribute{
				MarkdownDescription: "",
				Attributes: map[string]schema.Attribute{
					"connection_id": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
					},
					"object": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
					},
					"search_values": schema.StringAttribute{
						MarkdownDescription: "",
						CustomType:          jsontypes.NormalizedType{},
						Optional:            true,
						Computed:            true,
					},
					"configuration": schema.StringAttribute{
						MarkdownDescription: "",
						CustomType:          jsontypes.NormalizedType{},
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"new_name": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
					},
					"filter_logic": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
					},
				},
				Required: true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"fields": schema.SetNestedAttribute{
				MarkdownDescription: "",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"source": schema.SingleNestedAttribute{
							MarkdownDescription: "",
							Attributes: map[string]schema.Attribute{
								"model_id": schema.StringAttribute{
									MarkdownDescription: "",
									Required:            true,
								},
								"field": schema.StringAttribute{
									MarkdownDescription: "",
									Required:            true,
								},
							},
							Optional: true,
						},
						"target": schema.StringAttribute{
							MarkdownDescription: "",
							Required:            true,
						},
						"new": schema.BoolAttribute{
							MarkdownDescription: "",
							Optional:            true,
						},
						"override_value": schema.StringAttribute{
							MarkdownDescription: "",
							Optional:            true,
						},
						"sync_mode": schema.StringAttribute{
							MarkdownDescription: "",
							Optional:            true,
						},
						"encryption_enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the field should be encrypted",
							Optional:            true,
						},
					},
				},
				Required: true,
			},
			"override_fields": schema.SetNestedAttribute{
				MarkdownDescription: "",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"source": schema.SingleNestedAttribute{
							MarkdownDescription: "",
							Attributes: map[string]schema.Attribute{
								"model_id": schema.StringAttribute{
									MarkdownDescription: "",
									Required:            true,
								},
								"field": schema.StringAttribute{
									MarkdownDescription: "",
									Required:            true,
								},
							},
							Optional: true,
						},
						"target": schema.StringAttribute{
							MarkdownDescription: "",
							Required:            true,
						},
						"new": schema.BoolAttribute{
							MarkdownDescription: "",
							Optional:            true,
						},
						"override_value": schema.StringAttribute{
							MarkdownDescription: "",
							Optional:            true,
						},
						"sync_mode": schema.StringAttribute{
							MarkdownDescription: "",
							Optional:            true,
						},
					}},
				Optional: true,
			},
			"filters": schema.SetNestedAttribute{
				MarkdownDescription: "",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field_id": schema.StringAttribute{
							MarkdownDescription: "",
							Required:            true,
						},
						"field_type": schema.StringAttribute{
							MarkdownDescription: "",
							Required:            true,
						},
						"function": schema.StringAttribute{
							MarkdownDescription: "",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "",
							CustomType:          jsontypes.NormalizedType{},
							Optional:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: "",
							Optional:            true,
							Computed:            true,
						},
					}},
				Optional: true,
			},
			"filter_logic": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"overrides": schema.SetNestedAttribute{
				MarkdownDescription: "",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field_id": schema.StringAttribute{
							MarkdownDescription: "",
							Required:            true,
						},
						"function": schema.StringAttribute{
							MarkdownDescription: "",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "",
							CustomType:          jsontypes.NormalizedType{},
							Optional:            true,
						},
						"override": schema.StringAttribute{
							MarkdownDescription: "",
							CustomType:          jsontypes.NormalizedType{},
							Required:            true,
						}},
				},
				Optional: true,
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
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"hour": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"minute": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"month": schema.StringAttribute{
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"day_of_month": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"job_id": schema.Int64Attribute{
						MarkdownDescription: "",
						Optional:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"connection_id": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"run_after": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"sync_ids": schema.SetAttribute{
								MarkdownDescription: "",
								ElementType:         types.StringType,
								Optional:            true,
							},
							"bulk_sync_ids": schema.SetAttribute{
								MarkdownDescription: "",
								ElementType:         types.StringType,
								Optional:            true,
							},
						},
						Optional: true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
					},
					"run_after_success_only": schema.BoolAttribute{
						Optional: true,
					},
				},
				Required: true,
			},
			"identity": schema.SingleNestedAttribute{
				MarkdownDescription: "",
				Attributes: map[string]schema.Attribute{
					"source": schema.SingleNestedAttribute{
						MarkdownDescription: "",
						Attributes: map[string]schema.Attribute{
							"model_id": schema.StringAttribute{
								MarkdownDescription: "",
								Required:            true,
							},
							"field": schema.StringAttribute{
								MarkdownDescription: "",
								Required:            true,
							},
						},
						Required: true,
					},
					"target": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
					},
					"function": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
					},
					"remote_field_type_id": schema.StringAttribute{
						MarkdownDescription: "",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"new_field": schema.BoolAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Optional: true,
			},
			"sync_all_records": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"encryption_passphrase": schema.StringAttribute{
				MarkdownDescription: "Passphrase for encrypting sync data",
				Optional:            true,
				Sensitive:           true,
			},
			"only_enrich_updates": schema.BoolAttribute{
				MarkdownDescription: "Whether enrichment models only track changes",
				Optional:            true,
				Computed:            true,
			},
			"skip_initial_backfill": schema.BoolAttribute{
				MarkdownDescription: "Skip initial backfill, sync only new records",
				Optional:            true,
				Computed:            true,
			},
			"model_ids": schema.SetAttribute{
				MarkdownDescription: "Model IDs associated with this sync",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (r *syncResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sync"
}

type syncResourceResourceData struct {
	ID                   types.String `tfsdk:"id"`
	Organization         types.String `tfsdk:"organization"`
	Name                 types.String `tfsdk:"name"`
	Target               types.Object `tfsdk:"target"`
	Mode                 types.String `tfsdk:"mode"`
	Fields               types.Set    `tfsdk:"fields"`
	OverrideFields       types.Set    `tfsdk:"override_fields"`
	Filters              types.Set    `tfsdk:"filters"`
	FilterLogic          types.String `tfsdk:"filter_logic"`
	Overrides            types.Set    `tfsdk:"overrides"`
	Schedule             types.Object `tfsdk:"schedule"`
	Identity             types.Object `tfsdk:"identity"`
	SyncAllRecords       types.Bool   `tfsdk:"sync_all_records"`
	Active               types.Bool   `tfsdk:"active"`
	EncryptionPassphrase types.String `tfsdk:"encryption_passphrase"`
	OnlyEnrichUpdates    types.Bool   `tfsdk:"only_enrich_updates"`
	SkipInitialBackfill  types.Bool   `tfsdk:"skip_initial_backfill"`
	ModelIds             types.Set    `tfsdk:"model_ids"`
}

type Filter struct {
	FieldID   string               `json:"field_id" tfsdk:"field_id" mapstructure:"field_id"`
	FieldType string               `json:"field_type" tfsdk:"field_type" mapstructure:"field_type"`
	Function  string               `json:"function" tfsdk:"function" mapstructure:"function"`
	Value     jsontypes.Normalized `json:"value" tfsdk:"value" mapstructure:"value"`
	Label     string               `json:"label" tfsdk:"label" mapstructure:"label"`
}

func (Filter) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"field_id":   types.StringType,
		"field_type": types.StringType,
		"function":   types.StringType,
		"value":      jsontypes.NormalizedType{},
		"label":      types.StringType,
	}
}

type Target struct {
	ConnectionID  string               `json:"connection_id" tfsdk:"connection_id" mapstructure:"connection_id"`
	Object        *string              `json:"object" tfsdk:"object" mapstructure:"object"`
	SearchValues  jsontypes.Normalized `json:"search_values,omitempty" tfsdk:"search_values" mapstructure:"search_values,omitempty"`
	Configuration jsontypes.Normalized `json:"configuration,omitempty" tfsdk:"configuration" mapstructure:"configuration,omitempty"`
	NewName       *string              `json:"new_name,omitempty" tfsdk:"new_name" mapstructure:"new_name"`
	FilterLogic   *string              `json:"filter_logic,omitempty" tfsdk:"filter_logic" mapstructure:"filter_logic"`
}

func (Target) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"connection_id": types.StringType,
		"object":        types.StringType,
		"search_values": jsontypes.NormalizedType{},
		"configuration": jsontypes.NormalizedType{},
		"new_name":      types.StringType,
		"filter_logic":  types.StringType,
	}
}

type Override struct {
	FieldID  string               `json:"field_id" tfsdk:"field_id" mapstructure:"field_id"`
	Function string               `json:"function" tfsdk:"function" mapstructure:"function"`
	Value    jsontypes.Normalized `json:"value" tfsdk:"value" mapstructure:"value"`
	Override jsontypes.Normalized `json:"override" tfsdk:"override" mapstructure:"override"`
}

func (Override) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"field_id": types.StringType,
		"function": types.StringType,
		"value":    jsontypes.NormalizedType{},
		"override": jsontypes.NormalizedType{},
	}
}

type Schedule struct {
	Frequency           string       `tfsdk:"frequency"`
	DayOfWeek           *string      `tfsdk:"day_of_week"`
	Hour                *string      `tfsdk:"hour"`
	Minute              *string      `tfsdk:"minute"`
	Month               *string      `tfsdk:"month"`
	DayOfMonth          *string      `tfsdk:"day_of_month"`
	JobID               *int64       `tfsdk:"job_id"`
	ConnectionID        *string      `tfsdk:"connection_id"`
	RunAfter            types.Object `tfsdk:"run_after"`
	RunAfterSuccessOnly *bool        `tfsdk:"run_after_success_only"`
}

func (Schedule) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"frequency":     types.StringType,
		"day_of_week":   types.StringType,
		"hour":          types.StringType,
		"minute":        types.StringType,
		"month":         types.StringType,
		"day_of_month":  types.StringType,
		"job_id":        types.Int64Type,
		"connection_id": types.StringType,
		"run_after": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"sync_ids":      types.SetType{ElemType: types.StringType},
				"bulk_sync_ids": types.SetType{ElemType: types.StringType},
			},
		},
		"run_after_success_only": types.BoolType,
	}
}

type syncResource struct {
	provider *providerclient.Provider
}

func (r *syncResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *syncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data syncResourceResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var target Target
	diags = data.Target.As(ctx, &target, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	pt := &polytomic.Target{
		ConnectionId: target.ConnectionID,
		Object:       target.Object,
		NewName:      target.NewName,
		FilterLogic:  target.FilterLogic,
	}

	var searchValues map[string]interface{}
	if !target.SearchValues.IsNull() && !target.SearchValues.IsUnknown() {
		diags = target.SearchValues.Unmarshal(&searchValues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		pt.SearchValues = searchValues
	} else {
		pt.SearchValues = make(map[string]interface{})
	}

	var tConf map[string]interface{}
	if !target.Configuration.IsNull() && !target.Configuration.IsUnknown() {
		diags = target.Configuration.Unmarshal(&tConf)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		pt.Configuration = tConf
	} else {
		pt.Configuration = make(map[string]interface{})
	}

	var fields []*polytomic.ModelSyncField
	diags = data.Fields.ElementsAs(ctx, &fields, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var overrideFields []*polytomic.ModelSyncField
	diags = data.OverrideFields.ElementsAs(ctx, &overrideFields, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var filters []Filter
	diags = data.Filters.ElementsAs(ctx, &filters, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var pfilters []*polytomic.Filter
	for _, filter := range filters {
		f := &polytomic.Filter{
			FieldId:   pointer.To(filter.FieldID),
			FieldType: pointer.To(polytomic.FilterFieldReferenceType(filter.FieldType)),
			Function:  polytomic.FilterFunction(filter.Function),
			Label:     pointer.To(filter.Label),
		}

		var val interface{}
		if !filter.Value.IsNull() && !filter.Value.IsUnknown() {
			diags = filter.Value.Unmarshal(&val)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			f.Value = val
		}
		pfilters = append(pfilters, f)

	}

	var overrides []Override
	diags = data.Overrides.ElementsAs(ctx, &overrides, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var poverrides []*polytomic.Override
	for _, override := range overrides {
		o := &polytomic.Override{
			FieldId:  &override.FieldID,
			Function: pointer.To(polytomic.FilterFunction(override.Function)),
		}

		var val interface{}
		if !override.Value.IsNull() && !override.Value.IsUnknown() {
			diags = override.Value.Unmarshal(&val)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			o.Value = val
		}

		var ov interface{}
		if !override.Override.IsNull() && !override.Override.IsUnknown() {
			diags = override.Override.Unmarshal(&ov)
			if diags.HasError() {
				// if unmarshalling fails, try to use as string
				var ovStr string
				diags = override.Override.Unmarshal(&ovStr)
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}
				ov = ovStr
			}
			o.Override = ov
		}

		poverrides = append(poverrides, o)

	}

	var schedule polytomic.Schedule
	diags = data.Schedule.As(ctx, &schedule, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var identity polytomic.Identity
	diags = data.Identity.As(ctx, &identity, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	request := &polytomic.CreateModelSyncRequest{
		Name:                 data.Name.ValueString(),
		Target:               pt,
		Mode:                 polytomic.ModelSyncMode(data.Mode.ValueString()),
		Fields:               fields,
		OverrideFields:       overrideFields,
		Filters:              pfilters,
		Overrides:            poverrides,
		Schedule:             &schedule,
		EncryptionPassphrase: data.EncryptionPassphrase.ValueStringPointer(),
		OnlyEnrichUpdates:    data.OnlyEnrichUpdates.ValueBoolPointer(),
		SkipInitialBackfill:  data.SkipInitialBackfill.ValueBoolPointer(),
	}

	if !data.Organization.IsNull() {
		request.OrganizationId = data.Organization.ValueStringPointer()
	}
	if !data.FilterLogic.IsNull() {
		request.FilterLogic = data.FilterLogic.ValueStringPointer()
	}
	if !data.SyncAllRecords.IsNull() {
		request.SyncAllRecords = data.SyncAllRecords.ValueBoolPointer()
	}
	if !data.Active.IsNull() {
		request.Active = data.Active.ValueBoolPointer()
	}

	if identity.Source.ModelId != "" && identity.Source.Field != "" {
		request.Identity = &identity
	}
	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	sync, err := client.ModelSync.Create(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Error creating sync", err.Error())
		return
	}

	data, diags = syncDataFromResponse(ctx, sync.Data)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

func (r *syncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data syncResourceResourceData

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
	sync, err := client.ModelSync.Get(ctx, data.ID.ValueString())
	if err != nil {
		pErr := &ptcore.APIError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError("Error reading sync", err.Error())
		return
	}

	data, diags = syncDataFromResponse(ctx, sync.Data)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

func (r *syncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data syncResourceResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configuration syncResourceResourceData
	diags = req.Config.Get(ctx, &configuration)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var target Target
	diags = data.Target.As(ctx, &target, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	pt := &polytomic.Target{
		ConnectionId: target.ConnectionID,
		Object:       pointer.To(pointer.Get(target.Object)),
		NewName:      target.NewName,
		FilterLogic:  target.FilterLogic,
	}

	var searchValues map[string]interface{}
	if !target.SearchValues.IsNull() && !target.SearchValues.IsUnknown() {
		diags = target.SearchValues.Unmarshal(&searchValues)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		pt.SearchValues = searchValues
	} else {
		pt.SearchValues = make(map[string]interface{})
	}

	var tConf map[string]interface{}
	if !target.Configuration.IsNull() && !target.Configuration.IsUnknown() {
		diags = target.Configuration.Unmarshal(&tConf)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		pt.Configuration = tConf
	} else {
		pt.Configuration = make(map[string]interface{})
	}

	var fields []*polytomic.ModelSyncField
	diags = data.Fields.ElementsAs(ctx, &fields, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var overrideFields []*polytomic.ModelSyncField
	diags = data.OverrideFields.ElementsAs(ctx, &overrideFields, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var filters []*Filter
	diags = data.Filters.ElementsAs(ctx, &filters, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var pfilters []*polytomic.Filter
	for _, filter := range filters {
		f := &polytomic.Filter{
			FieldId:   pointer.To(filter.FieldID),
			FieldType: pointer.To(polytomic.FilterFieldReferenceType(filter.FieldType)),
			Function:  polytomic.FilterFunction(filter.Function),
			Label:     pointer.To(filter.Label),
		}

		var val interface{}
		if !filter.Value.IsNull() && !filter.Value.IsUnknown() {
			diags = filter.Value.Unmarshal(&val)
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			f.Value = val
		}
		pfilters = append(pfilters, f)

	}

	var overrides []Override
	diags = data.Overrides.ElementsAs(ctx, &overrides, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var poverrides []*polytomic.Override
	for _, override := range overrides {
		o := &polytomic.Override{
			FieldId:  pointer.To(override.FieldID),
			Function: pointer.To(polytomic.FilterFunction(override.Function)),
		}

		var val interface{}
		if !override.Value.IsNull() && !override.Value.IsUnknown() {
			diags = override.Value.Unmarshal(&val)
			if diags.HasError() {
				// if unmarshalling fails, try to use as string
				var valStr string
				diags = override.Value.Unmarshal(&valStr)
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}
				val = valStr
			}
			o.Value = val
		}

		var ov interface{}
		if !override.Override.IsNull() && !override.Override.IsUnknown() {
			diags = override.Override.Unmarshal(&ov)
			if diags.HasError() {
				// if unmarshalling fails, try to use as string
				var ovStr string
				diags = override.Override.Unmarshal(&ovStr)
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}
				ov = ovStr
			}
			o.Override = ov
		}

		poverrides = append(poverrides, o)

	}

	var schedule polytomic.Schedule
	diags = data.Schedule.As(ctx, &schedule, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var identity *polytomic.Identity
	diags = data.Identity.As(ctx, &identity, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    false,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	request := &polytomic.UpdateModelSyncRequest{
		Name:                 data.Name.ValueString(),
		Target:               pt,
		Mode:                 polytomic.ModelSyncMode(data.Mode.ValueString()),
		Fields:               fields,
		OverrideFields:       overrideFields,
		Filters:              pfilters,
		Overrides:            poverrides,
		Schedule:             &schedule,
		Identity:             identity,
		EncryptionPassphrase: data.EncryptionPassphrase.ValueStringPointer(),
		OnlyEnrichUpdates:    data.OnlyEnrichUpdates.ValueBoolPointer(),
		SkipInitialBackfill:  data.SkipInitialBackfill.ValueBoolPointer(),
	}

	if !data.Organization.IsNull() {
		request.OrganizationId = data.Organization.ValueStringPointer()
	}
	if !data.FilterLogic.IsNull() {
		request.FilterLogic = data.FilterLogic.ValueStringPointer()
	}
	if !data.SyncAllRecords.IsNull() {
		request.SyncAllRecords = data.SyncAllRecords.ValueBoolPointer()
	}
	if !data.Active.IsNull() {
		request.Active = data.Active.ValueBoolPointer()
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	sync, err := client.ModelSync.Update(ctx, data.ID.ValueString(), request)
	if err != nil {
		resp.Diagnostics.AddError("Error updating sync", err.Error())
		return
	}

	data, diags = syncDataFromResponse(ctx, sync.Data)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *syncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data syncResourceResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	err = client.ModelSync.Remove(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting sync", err.Error())
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *syncResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// syncDataFromResponse converts a Polytomic API response to Terraform resource data.
// This is the single source of truth for all CRUD operations.
func syncDataFromResponse(ctx context.Context, sync *polytomic.ModelSyncResponse) (syncResourceResourceData, diag.Diagnostics) {
	var data syncResourceResourceData
	var diags diag.Diagnostics

	// Basic fields
	data.ID = types.StringPointerValue(sync.Id)
	data.Organization = types.StringPointerValue(sync.OrganizationId)
	data.Name = types.StringPointerValue(sync.Name)
	data.Mode = types.StringValue(string(pointer.Get(sync.Mode)))
	data.Active = types.BoolPointerValue(sync.Active)
	data.SyncAllRecords = types.BoolPointerValue(sync.SyncAllRecords)
	data.OnlyEnrichUpdates = types.BoolPointerValue(sync.OnlyEnrichUpdates)
	data.SkipInitialBackfill = types.BoolPointerValue(sync.SkipInitialBackfill)

	// Target - using jsontypes for SearchValues and Configuration
	searchValJSON, err := json.Marshal(sync.Target.SearchValues)
	if err != nil {
		diags.AddError("Error marshaling search values", err.Error())
		return data, diags
	}
	confJSON, err := json.Marshal(sync.Target.Configuration)
	if err != nil {
		diags.AddError("Error marshaling configuration", err.Error())
		return data, diags
	}

	var searchValNormalized jsontypes.Normalized
	if string(searchValJSON) == "null" {
		searchValNormalized = jsontypes.NewNormalizedNull()
	} else {
		searchValNormalized = jsontypes.NewNormalizedValue(string(searchValJSON))
	}

	var confNormalized jsontypes.Normalized
	if string(confJSON) == "null" {
		confNormalized = jsontypes.NewNormalizedNull()
	} else {
		confNormalized = jsontypes.NewNormalizedValue(string(confJSON))
	}

	targetData := Target{
		ConnectionID:  sync.Target.ConnectionId,
		Object:        sync.Target.Object,
		SearchValues:  searchValNormalized,
		Configuration: confNormalized,
		NewName:       sync.Target.NewName,
		FilterLogic:   sync.Target.FilterLogic,
	}
	data.Target, diags = types.ObjectValueFrom(ctx, Target{}.AttrTypes(), targetData)
	if diags.HasError() {
		return data, diags
	}

	// Fields
	data.Fields, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"source": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"model_id": types.StringType,
					"field":    types.StringType,
				}},
			"target":             types.StringType,
			"new":                types.BoolType,
			"override_value":     types.StringType,
			"sync_mode":          types.StringType,
			"encryption_enabled": types.BoolType,
		}}, sync.Fields)
	if diags.HasError() {
		return data, diags
	}

	// Override Fields
	data.OverrideFields, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"source": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"model_id": types.StringType,
					"field":    types.StringType,
				}},
			"target":         types.StringType,
			"new":            types.BoolType,
			"override_value": types.StringType,
			"sync_mode":      types.StringType,
		}}, sync.OverrideFields)
	if diags.HasError() {
		return data, diags
	}

	// FilterLogic
	if sync.FilterLogic != nil {
		data.FilterLogic = types.StringPointerValue(sync.FilterLogic)
	} else {
		data.FilterLogic = types.StringNull()
	}

	// Filters - convert SDK filters to TF filters
	var tfFilters []Filter
	for _, f := range sync.Filters {
		var valNormalized jsontypes.Normalized
		valJSON, err := json.Marshal(f.Value)
		if err != nil {
			diags.AddError("Error marshaling filter value", err.Error())
			return data, diags
		}
		if string(valJSON) == "null" {
			valNormalized = jsontypes.NewNormalizedNull()
		} else {
			valNormalized = jsontypes.NewNormalizedValue(string(valJSON))
		}

		tfFilters = append(tfFilters, Filter{
			FieldID:   pointer.Get(f.FieldId),
			FieldType: string(pointer.Get(f.FieldType)),
			Function:  string(f.Function),
			Value:     valNormalized,
			Label:     pointer.GetString(f.Label),
		})
	}
	data.Filters, diags = types.SetValueFrom(ctx, types.ObjectType{AttrTypes: Filter{}.AttrTypes()}, tfFilters)
	if diags.HasError() {
		return data, diags
	}

	// Overrides - convert SDK overrides to TF overrides
	var tfOverrides []Override
	for _, o := range sync.Overrides {
		var valNormalized jsontypes.Normalized
		valJSON, err := json.Marshal(o.Value)
		if err != nil {
			diags.AddError("Error marshaling override value", err.Error())
			return data, diags
		}
		if string(valJSON) == "null" {
			valNormalized = jsontypes.NewNormalizedNull()
		} else {
			valNormalized = jsontypes.NewNormalizedValue(string(valJSON))
		}

		var overrideNormalized jsontypes.Normalized
		// Handle both string and complex override values
		if v, ok := o.Override.(string); ok {
			overrideNormalized = jsontypes.NewNormalizedValue(v)
		} else {
			overrideJSON, err := json.Marshal(o.Override)
			if err != nil {
				diags.AddError("Error marshaling override override", err.Error())
				return data, diags
			}
			if string(overrideJSON) == "null" {
				overrideNormalized = jsontypes.NewNormalizedNull()
			} else {
				overrideNormalized = jsontypes.NewNormalizedValue(string(overrideJSON))
			}
		}

		tfOverrides = append(tfOverrides, Override{
			FieldID:  pointer.Get(o.FieldId),
			Function: string(pointer.Get(o.Function)),
			Value:    valNormalized,
			Override: overrideNormalized,
		})
	}
	data.Overrides, diags = types.SetValueFrom(ctx, types.ObjectType{AttrTypes: Override{}.AttrTypes()}, tfOverrides)
	if diags.HasError() {
		return data, diags
	}

	// Schedule
	data.Schedule, diags = types.ObjectValueFrom(ctx, Schedule{}.AttrTypes(), sync.Schedule)
	if diags.HasError() {
		return data, diags
	}

	// Identity
	data.Identity, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"source": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"model_id": types.StringType,
				"field":    types.StringType,
			},
		},
		"target":               types.StringType,
		"function":             types.StringType,
		"remote_field_type_id": types.StringType,
		"new_field":            types.BoolType,
	}, sync.Identity)
	if diags.HasError() {
		return data, diags
	}

	// ModelIds - extract unique model IDs from fields
	modelIDMap := make(map[string]bool)
	for _, field := range sync.Fields {
		if field.Source != nil && field.Source.ModelId != "" {
			modelIDMap[field.Source.ModelId] = true
		}
	}
	modelIDs := make([]string, 0, len(modelIDMap))
	for id := range modelIDMap {
		modelIDs = append(modelIDs, id)
	}
	data.ModelIds, diags = types.SetValueFrom(ctx, types.StringType, modelIDs)
	if diags.HasError() {
		return data, diags
	}

	return data, diags
}
