package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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
						Type:                types.MapType{ElemType: types.StringType},
						Optional:            true,
					},
					"configuration": {
						MarkdownDescription: "",
						Type:                types.MapType{ElemType: types.StringType},
						Optional:            true,
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
						Required:            true,
					},
					"label": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
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
				Type: types.SetType{ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"field_id": types.StringType,
						"function": types.StringType,
						"value":    types.StringType,
					}},
				},
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

	var target polytomic.Target
	diags = data.Target.As(ctx, &target, types.ObjectAsOptions{
		UnhandledNullAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
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

	var filters []polytomic.Filter
	diags = data.Filters.ElementsAs(ctx, &filters, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var overrides []polytomic.Override
	diags = data.Overrides.ElementsAs(ctx, &overrides, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
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
		Target:         target,
		Mode:           data.Mode.ValueString(),
		Fields:         fields,
		OverrideFields: overrideFields,
		Filters:        filters,
		FilterLogic:    data.FilterLogic.ValueString(),
		Overrides:      overrides,
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

	data.Target, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"connection_id": types.StringType,
		"object":        types.StringType,
		"search_values": types.MapType{ElemType: types.StringType},
		"configuration": types.MapType{ElemType: types.StringType},
		"new_name":      types.StringType,
		"filter_logic":  types.StringType,
	}, sync.Target)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
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
	data.Filters, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"field_id":   types.StringType,
			"field_type": types.StringType,
			"function":   types.StringType,
			"value":      types.StringType,
			"label":      types.StringType,
		}}, sync.Filters)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	if sync.FilterLogic != "" {
		data.FilterLogic = types.StringValue(sync.FilterLogic)
	}
	data.Overrides, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"field_id": types.StringType,
			"function": types.StringType,
			"value":    types.StringType,
		}}, sync.Overrides)
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

	data.Target, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"connection_id": types.StringType,
		"object":        types.StringType,
		"search_values": types.MapType{ElemType: types.StringType},
		"configuration": types.MapType{ElemType: types.StringType},
		"new_name":      types.StringType,
		"filter_logic":  types.StringType,
	}, sync.Target)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
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
	data.Filters, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"field_id":   types.StringType,
			"field_type": types.StringType,
			"function":   types.StringType,
			"value":      types.StringType,
			"label":      types.StringType,
		}}, sync.Filters)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if sync.FilterLogic != "" {
		data.FilterLogic = types.StringValue(sync.FilterLogic)
	} else {
		data.FilterLogic = types.StringNull()
	}

	data.Overrides, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"field_id": types.StringType,
			"function": types.StringType,
			"value":    types.StringType,
		}}, sync.Overrides)
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

	var target polytomic.Target
	diags = data.Target.As(ctx, &target, types.ObjectAsOptions{
		UnhandledNullAsEmpty: true,
	})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
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

	var filters []polytomic.Filter
	diags = data.Filters.ElementsAs(ctx, &filters, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var overrides []polytomic.Override
	diags = data.Overrides.ElementsAs(ctx, &overrides, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
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
		Target:         target,
		Mode:           data.Mode.ValueString(),
		Fields:         fields,
		OverrideFields: overrideFields,
		Filters:        filters,
		FilterLogic:    data.FilterLogic.ValueString(),
		Overrides:      overrides,
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

	var plannedTarget polytomic.Target
	diags = configuration.Target.As(ctx, &plannedTarget, types.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if plannedTarget.Configuration != nil &&
		sync.Target.Configuration == nil {
		sync.Target.Configuration = plannedTarget.Configuration
	}

	data.Target, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"connection_id": types.StringType,
		"object":        types.StringType,
		"search_values": types.MapType{ElemType: types.StringType},
		"configuration": types.MapType{ElemType: types.StringType},
		"new_name":      types.StringType,
		"filter_logic":  types.StringType,
	}, sync.Target)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
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
	data.Filters, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"field_id":   types.StringType,
			"field_type": types.StringType,
			"function":   types.StringType,
			"value":      types.StringType,
			"label":      types.StringType,
		}}, sync.Filters)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if configuration.FilterLogic.IsNull() {
		data.FilterLogic = configuration.FilterLogic
	} else {
		data.FilterLogic = types.StringValue(sync.FilterLogic)
	}

	data.Overrides, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"field_id": types.StringType,
			"function": types.StringType,
			"value":    types.StringType,
		}}, sync.Overrides)
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
