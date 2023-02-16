package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &syncResource{}
var _ resource.ResourceWithImportState = &syncResource{}

func (r *syncResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: ":meta:subcategory:Syncs: Sync",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"organization": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"name": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"target": {
				MarkdownDescription: "",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"connection_id": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
					"object": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
					},
					"search_values": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
					},
					"configuration": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
					},
					"new_name": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
					"filter_logic": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
				}),
				Required: true,
			},
			"mode": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"fields": {
				MarkdownDescription: "",
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"source": {
						MarkdownDescription: "",
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"model_id": {
								MarkdownDescription: "",
								Type:                types.StringType,
								Required:            true,
							},
							"field": {
								MarkdownDescription: "",
								Type:                types.StringType,
								Required:            true,
							},
						}),
						Required: true,
					},
					"target": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
					"new": {
						MarkdownDescription: "",
						Type:                types.BoolType,
						Optional:            true,
					},
					"override_value": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
					"sync_mode": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
				}),
				Required: true,
			},
			"override_fields": {
				MarkdownDescription: "",
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"source": {
						MarkdownDescription: "",
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"model_id": {
								MarkdownDescription: "",
								Type:                types.StringType,
								Required:            true,
							},
							"field": {
								MarkdownDescription: "",
								Type:                types.StringType,
								Required:            true,
							},
						}),
						Required: true,
					},
					"target": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
					"new": {
						MarkdownDescription: "",
						Type:                types.BoolType,
						Optional:            true,
					},
					"override_value": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
					"sync_mode": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
				}),
				Optional: true,
			},
			"filters": {
				MarkdownDescription: "",
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"field_id": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
					"field_type": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
					"function": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
					"value": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
					"label": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
				}),
				Optional: true,
			},
			"filter_logic": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Optional:            true,
			},
			"overrides": {
				MarkdownDescription: "",
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"field_id": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
					"function": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
					"value": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
					"override": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
				}),
				Optional: true,
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
					},
					"hour": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
					"minute": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
					"month": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
					"day_of_month": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
					},
				}),
				Required: true,
			},
			"identity": {
				MarkdownDescription: "",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"source": {
						MarkdownDescription: "",
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"model_id": {
								MarkdownDescription: "",
								Type:                types.StringType,
								Required:            true,
							},
							"field": {
								MarkdownDescription: "",
								Type:                types.StringType,
								Required:            true,
							},
						}),
						Required: true,
					},
					"target": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
					"function": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
					},
					"remote_field_type_id": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
					},
					"new_field": {
						MarkdownDescription: "",
						Type:                types.BoolType,
						Optional:            true,
						Computed:            true,
					},
				}),
				Optional: true,
			},
			"sync_all_records": {
				MarkdownDescription: "",
				Type:                types.BoolType,
				Optional:            true,
				Computed:            true,
			},
		},
	}, nil
}

func (r *syncResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *syncResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sync"
}

type syncResourceResourceData struct {
	ID             types.String `tfsdk:"id"`
	Organization   types.String `tfsdk:"organization"`
	Name           types.String `tfsdk:"name"`
	Target         types.Object `tfsdk:"target"`
	Mode           types.String `tfsdk:"mode"`
	Fields         types.Set    `tfsdk:"fields"`
	OverrideFields types.Set    `tfsdk:"override_fields"`
	Filters        types.Set    `tfsdk:"filters"`
	FilterLogic    types.String `tfsdk:"filter_logic"`
	Overrides      types.Set    `tfsdk:"overrides"`
	Schedule       types.Object `tfsdk:"schedule"`
	Identity       types.Object `tfsdk:"identity"`
	SyncAllRecords types.Bool   `tfsdk:"sync_all_records"`
}

type Filter struct {
	FieldID   string `json:"field_id" tfsdk:"field_id" mapstructure:"field_id"`
	FieldType string `json:"field_type" tfsdk:"field_type" mapstructure:"field_type"`
	Function  string `json:"function" tfsdk:"function" mapstructure:"function"`
	Value     string `json:"value" tfsdk:"value" mapstructure:"value"`
	Label     string `json:"label" tfsdk:"label" mapstructure:"label"`
}

type Target struct {
	ConnectionID  string  `json:"connection_id" tfsdk:"connection_id" mapstructure:"connection_id"`
	Object        *string `json:"object" tfsdk:"object" mapstructure:"object"`
	SearchValues  string  `json:"search_values,omitempty" tfsdk:"search_values" mapstructure:"search_values,omitempty"`
	Configuration string  `json:"configuration,omitempty" tfsdk:"configuration" mapstructure:"configuration,omitempty"`
	NewName       *string `json:"new_name,omitempty" tfsdk:"new_name" mapstructure:"new_name"`
	FilterLogic   *string `json:"filter_logic,omitempty" tfsdk:"filter_logic" mapstructure:"filter_logic"`
}

type Override struct {
	FieldID  string  `json:"field_id" tfsdk:"field_id" mapstructure:"field_id"`
	Function string  `json:"function" tfsdk:"function" mapstructure:"function"`
	Value    *string `json:"value" tfsdk:"value" mapstructure:"value"`
	Override *string `json:"override" tfsdk:"override" mapstructure:"override"`
}

type syncResource struct {
	client *polytomic.Client
}

func (r *syncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data syncResourceResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var target Target
	diags = data.Target.As(ctx, &target, types.ObjectAsOptions{
		UnhandledNullAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	pt := polytomic.Target{
		ConnectionID: target.ConnectionID,
		Object:       target.Object,
		NewName:      target.NewName,
		FilterLogic:  target.FilterLogic,
	}

	if target.SearchValues != "" {
		var searchValues map[string]interface{}
		err := json.Unmarshal([]byte(target.SearchValues), &searchValues)
		if err != nil {
			resp.Diagnostics.AddError("Error unmarshalling search values", err.Error())
			return
		}
		pt.SearchValues = searchValues
	} else {
		pt.SearchValues = make(map[string]interface{})
	}

	if target.Configuration != "" {
		var configuration map[string]interface{}
		err := json.Unmarshal([]byte(target.Configuration), &configuration)
		if err != nil {
			resp.Diagnostics.AddError("Error unmarshalling configuration", err.Error())
			return
		}
		pt.Configuration = configuration
	} else {
		pt.Configuration = make(map[string]interface{})
	}

	var fields []polytomic.SyncField
	diags = data.Fields.ElementsAs(ctx, &fields, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var overrideFields []polytomic.SyncField
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

	var pfilters []polytomic.Filter
	for _, filter := range filters {
		f := polytomic.Filter{
			FieldID:   filter.FieldID,
			FieldType: filter.FieldType,
			Function:  filter.Function,
			Label:     filter.Label,
		}

		var val interface{}
		if filter.Value != "" {
			err := json.Unmarshal([]byte(filter.Value), &val)
			if err != nil {
				resp.Diagnostics.AddError("Failed to unmarshal filter value", err.Error())
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

	var poverrides []polytomic.Override
	for _, override := range overrides {
		o := polytomic.Override{
			FieldID:  override.FieldID,
			Function: override.Function,
		}

		var val interface{}
		if override.Value != nil {
			err := json.Unmarshal([]byte(*override.Value), &val)
			if err != nil {
				resp.Diagnostics.AddError("Failed to unmarshal override value", err.Error())
				return
			}
			o.Value = val
		}

		var ov interface{}
		if override.Override != nil {
			err := json.Unmarshal([]byte(*override.Override), &ov)
			if err != nil {
				resp.Diagnostics.AddError("Failed to unmarshal override override", err.Error())
				return
			}
			o.Override = ov
		}

		poverrides = append(poverrides, o)

	}

	var schedule polytomic.Schedule
	diags = data.Schedule.As(ctx, &schedule, types.ObjectAsOptions{
		UnhandledNullAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var identity polytomic.Identity
	diags = data.Identity.As(ctx, &identity, types.ObjectAsOptions{
		UnhandledNullAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	request := polytomic.SyncRequest{
		Name:           data.Name.ValueString(),
		OrganizationID: data.Organization.ValueString(),
		Target:         pt,
		Mode:           data.Mode.ValueString(),
		Fields:         fields,
		OverrideFields: overrideFields,
		Filters:        pfilters,
		FilterLogic:    data.FilterLogic.ValueString(),
		Overrides:      poverrides,
		Schedule:       schedule,
		SyncAllRecords: data.SyncAllRecords.ValueBool(),
	}

	if identity.Source.ModelID != "" && identity.Source.Field != "" {
		request.Identity = &identity
	}

	sync, err := r.client.Syncs().Create(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Error creating sync", err.Error())
		return
	}

	data.ID = types.StringValue(sync.ID)
	data.Organization = types.StringValue(sync.OrganizationID)
	data.Name = types.StringValue(sync.Name)

	t := Target{
		ConnectionID: sync.Target.ConnectionID,
		Object:       sync.Target.Object,
		NewName:      sync.Target.NewName,
		FilterLogic:  sync.Target.FilterLogic,
	}

	sval, err := json.Marshal(sync.Target.SearchValues)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling search values", err.Error())
		return
	}
	t.SearchValues = string(sval)

	tval, err := json.Marshal(sync.Target.Configuration)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling configuration", err.Error())
		return
	}
	t.Configuration = string(tval)

	data.Target, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"connection_id": types.StringType,
		"object":        types.StringType,
		"search_values": types.StringType,
		"configuration": types.StringType,
		"new_name":      types.StringType,
		"filter_logic":  types.StringType,
	}, t)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if string(sval) == "null" {
		data.Target.Attributes()["search_values"] = types.StringNull()
	}

	if string(tval) == "null" {
		data.Target.Attributes()["configuration"] = types.StringNull()
	}

	data.Mode = types.StringValue(sync.Mode)
	data.Fields, diags = types.SetValueFrom(ctx, types.ObjectType{
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
		}}, sync.Fields)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
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
		resp.Diagnostics.Append(diags...)
		return
	}

	if sync.FilterLogic != "" {
		data.FilterLogic = types.StringValue(sync.FilterLogic)
	}

	var resOverrides []Override
	for _, o := range sync.Overrides {
		res := Override{
			FieldID:  o.FieldID,
			Function: o.Function,
		}
		val, err := json.Marshal(o.Value)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling override value", err.Error())
			return
		}

		if string(val) == "null" {
			res.Value = nil
		} else {
			res.Value = pointer.ToString(string(val))
		}
		oval, err := json.Marshal(o.Override)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling override override", err.Error())
			return
		}

		if string(oval) == "null" {
			res.Override = nil
		} else {
			res.Override = pointer.ToString(string(oval))
		}
		resOverrides = append(resOverrides, res)
	}

	data.Overrides, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"field_id": types.StringType,
			"function": types.StringType,
			"value":    types.StringType,
			"override": types.StringType,
		}}, resOverrides)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Schedule, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"frequency":    types.StringType,
		"day_of_week":  types.StringType,
		"hour":         types.StringType,
		"minute":       types.StringType,
		"month":        types.StringType,
		"day_of_month": types.StringType,
	}, sync.Schedule)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
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
		resp.Diagnostics.Append(diags...)
		return
	}
	data.SyncAllRecords = types.BoolValue(sync.SyncAllRecords)

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

	sync, err := r.client.Syncs().Get(ctx, data.ID.ValueString())
	if err != nil {
		pErr := polytomic.ApiError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError("Error reading sync", err.Error())
		return
	}

	data.ID = types.StringValue(sync.ID)
	data.Organization = types.StringValue(sync.OrganizationID)
	data.Name = types.StringValue(sync.Name)

	t := Target{
		ConnectionID: sync.Target.ConnectionID,
		Object:       sync.Target.Object,
		NewName:      sync.Target.NewName,
		FilterLogic:  sync.Target.FilterLogic,
	}

	sval, err := json.Marshal(sync.Target.SearchValues)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling search values", err.Error())
		return
	}
	t.SearchValues = string(sval)

	tval, err := json.Marshal(sync.Target.Configuration)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling configuration", err.Error())
		return
	}
	t.Configuration = string(tval)

	data.Target, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"connection_id": types.StringType,
		"object":        types.StringType,
		"search_values": types.StringType,
		"configuration": types.StringType,
		"new_name":      types.StringType,
		"filter_logic":  types.StringType,
	}, t)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if string(sval) == "null" {
		data.Target.Attributes()["search_values"] = types.StringNull()
	}

	if string(tval) == "null" {
		data.Target.Attributes()["configuration"] = types.StringNull()
	}

	data.Mode = types.StringValue(sync.Mode)
	data.Fields, diags = types.SetValueFrom(ctx, types.ObjectType{
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
		}}, sync.Fields)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
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
		resp.Diagnostics.Append(diags...)
		return
	}

	if sync.FilterLogic != "" {
		data.FilterLogic = types.StringValue(sync.FilterLogic)
	} else {
		data.FilterLogic = types.StringNull()
	}

	var resOverrides []Override
	for _, o := range sync.Overrides {
		res := Override{
			FieldID:  o.FieldID,
			Function: o.Function,
		}
		val, err := json.Marshal(o.Value)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling override value", err.Error())
			return
		}
		if string(val) == "null" {
			res.Value = nil
		} else {
			res.Value = pointer.ToString(string(val))
		}
		oval, err := json.Marshal(o.Override)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling override override", err.Error())
			return
		}
		if string(oval) == "null" {
			res.Override = nil
		} else {
			res.Override = pointer.ToString(string(oval))
		}
		resOverrides = append(resOverrides, res)
	}

	data.Overrides, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"field_id": types.StringType,
			"function": types.StringType,
			"value":    types.StringType,
			"override": types.StringType,
		}}, resOverrides)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Schedule, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"frequency":    types.StringType,
		"day_of_week":  types.StringType,
		"hour":         types.StringType,
		"minute":       types.StringType,
		"month":        types.StringType,
		"day_of_month": types.StringType,
	}, sync.Schedule)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
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
		resp.Diagnostics.Append(diags...)
		return
	}
	data.SyncAllRecords = types.BoolValue(sync.SyncAllRecords)

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
	diags = data.Target.As(ctx, &target, types.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	pt := polytomic.Target{
		ConnectionID: target.ConnectionID,
		Object:       target.Object,
		NewName:      target.NewName,
		FilterLogic:  target.FilterLogic,
	}

	if target.SearchValues != "" {
		var searchValues map[string]interface{}
		err := json.Unmarshal([]byte(target.SearchValues), &searchValues)
		if err != nil {
			resp.Diagnostics.AddError("Error unmarshalling search values", err.Error())
			return
		}
		pt.SearchValues = searchValues
	} else {
		pt.SearchValues = make(map[string]interface{})
	}

	if target.Configuration != "" {
		var configuration map[string]interface{}
		err := json.Unmarshal([]byte(target.Configuration), &configuration)
		if err != nil {
			resp.Diagnostics.AddError("Error unmarshalling configuration", err.Error())
			return
		}
		pt.Configuration = configuration
	} else {
		pt.Configuration = make(map[string]interface{})
	}

	var fields []polytomic.SyncField
	diags = data.Fields.ElementsAs(ctx, &fields, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var overrideFields []polytomic.SyncField
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

	var pfilters []polytomic.Filter
	for _, filter := range filters {
		f := polytomic.Filter{
			FieldID:   filter.FieldID,
			FieldType: filter.FieldType,
			Function:  filter.Function,
			Label:     filter.Label,
		}

		var val interface{}
		if filter.Value != "" {
			err := json.Unmarshal([]byte(filter.Value), &val)
			if err != nil {
				resp.Diagnostics.AddError("Failed to unmarshal filter value", err.Error())
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

	var poverrides []polytomic.Override
	for _, override := range overrides {
		o := polytomic.Override{
			FieldID:  override.FieldID,
			Function: override.Function,
		}

		var val interface{}
		if override.Value != nil {
			err := json.Unmarshal([]byte(*override.Value), &val)
			if err != nil {
				resp.Diagnostics.AddError("Failed to unmarshal override value", err.Error())
				return
			}
			o.Value = val
		}

		var ov interface{}
		if override.Override != nil {
			err := json.Unmarshal([]byte(*override.Override), &ov)
			if err != nil {
				resp.Diagnostics.AddError("Failed to unmarshal override override", err.Error())
				return
			}
			o.Override = ov
		}

		poverrides = append(poverrides, o)

	}

	var schedule polytomic.Schedule
	diags = data.Schedule.As(ctx, &schedule, types.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var identity *polytomic.Identity
	diags = data.Identity.As(ctx, &identity, types.ObjectAsOptions{
		UnhandledNullAsEmpty: false,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	request := polytomic.SyncRequest{
		Name:           data.Name.ValueString(),
		OrganizationID: data.Organization.ValueString(),
		Target:         pt,
		Mode:           data.Mode.ValueString(),
		Fields:         fields,
		OverrideFields: overrideFields,
		Filters:        pfilters,
		FilterLogic:    data.FilterLogic.ValueString(),
		Overrides:      poverrides,
		Schedule:       schedule,
		Identity:       identity,
		SyncAllRecords: data.SyncAllRecords.ValueBool(),
	}
	sync, err := r.client.Syncs().Update(ctx, data.ID.ValueString(), request)
	if err != nil {
		resp.Diagnostics.AddError("Error creating sync", err.Error())
		return
	}

	data.ID = types.StringValue(sync.ID)
	data.Organization = types.StringValue(sync.OrganizationID)
	data.Name = types.StringValue(sync.Name)

	t := Target{
		ConnectionID: sync.Target.ConnectionID,
		Object:       sync.Target.Object,
		NewName:      sync.Target.NewName,
		FilterLogic:  sync.Target.FilterLogic,
	}

	sval, err := json.Marshal(sync.Target.SearchValues)
	if err != nil {
		resp.Diagnostics.AddError("Error marshalling search values", err.Error())
		return
	}
	t.SearchValues = string(sval)

	tval, err := json.Marshal(sync.Target.Configuration)
	if err != nil {
		resp.Diagnostics.AddError("Error marshalling target configuration", err.Error())
		return
	}
	t.Configuration = string(tval)

	data.Target, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"connection_id": types.StringType,
		"object":        types.StringType,
		"search_values": types.StringType,
		"configuration": types.StringType,
		"new_name":      types.StringType,
		"filter_logic":  types.StringType,
	}, t)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if string(sval) == "null" {
		data.Target.Attributes()["search_values"] = types.StringNull()
	}

	if string(tval) == "null" {
		data.Target.Attributes()["configuration"] = types.StringNull()
	}

	data.Mode = types.StringValue(sync.Mode)
	data.Fields, diags = types.SetValueFrom(ctx, types.ObjectType{
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
		}}, sync.Fields)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
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
		resp.Diagnostics.Append(diags...)
		return
	}

	if configuration.FilterLogic.IsNull() {
		data.FilterLogic = configuration.FilterLogic
	} else {
		data.FilterLogic = types.StringValue(sync.FilterLogic)
	}

	var resOverrides []Override
	for _, o := range sync.Overrides {
		res := Override{
			FieldID:  o.FieldID,
			Function: o.Function,
		}
		val, err := json.Marshal(o.Value)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling override value", err.Error())
			return
		}

		if string(val) == "null" {
			res.Value = nil
		} else {
			res.Value = pointer.ToString(string(val))
		}
		oval, err := json.Marshal(o.Override)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling override override", err.Error())
			return
		}

		if string(oval) == "null" {
			res.Override = nil
		} else {
			res.Override = pointer.ToString(string(oval))
		}
		resOverrides = append(resOverrides, res)
	}
	data.Overrides, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"field_id": types.StringType,
			"function": types.StringType,
			"value":    types.StringType,
			"override": types.StringType,
		}}, resOverrides)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Schedule, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"frequency":    types.StringType,
		"day_of_week":  types.StringType,
		"hour":         types.StringType,
		"minute":       types.StringType,
		"month":        types.StringType,
		"day_of_month": types.StringType,
	}, sync.Schedule)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
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
		resp.Diagnostics.Append(diags...)
		return
	}
	data.SyncAllRecords = types.BoolValue(sync.SyncAllRecords)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

func (r *syncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data syncResourceResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	err := r.client.Syncs().Delete(ctx, data.ID.ValueString())
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
