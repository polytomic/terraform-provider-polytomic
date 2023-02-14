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
var _ resource.Resource = &modelResource{}
var _ resource.ResourceWithImportState = &modelResource{}

func (r *modelResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: ":meta:subcategory:Models: Model",
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
			"connection_id": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"name": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"type": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Computed:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"version": {
				MarkdownDescription: "",
				Type:                types.Int64Type,
				Computed:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"configuration": {
				MarkdownDescription: "",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Required: true,
			},
			"fields": {
				MarkdownDescription: "",
				Type:                types.SetType{ElemType: types.StringType},
				Optional:            true,
				Computed:            true,
			},
			"additional_fields": {
				MarkdownDescription: "",
				Type: types.SetType{ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"type":  types.StringType,
						"label": types.StringType,
					},
				}},
				Optional: true,
				Computed: true,
			},
			"relations": {
				MarkdownDescription: "",
				Type: types.SetType{ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"to": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"model_id": types.StringType,
								"field":    types.StringType,
							},
						},
						"from": types.StringType,
					},
				}},
				Optional: true,
				Computed: true,
			},
			"identifier": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"tracking_columns": {
				MarkdownDescription: "",
				Type:                types.SetType{ElemType: types.StringType},
				Optional:            true,
				Computed:            true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
		},
	}, nil
}

func (r *modelResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || !req.State.Raw.IsKnown() {
		return
	}
	config := &modelResourceResourceData{}
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)

	plan := &modelResourceResourceData{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Identifier.IsNull() {
		plan.Identifier = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)

}

func (r *modelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *modelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

type modelResourceResourceData struct {
	ID               types.String `tfsdk:"id"`
	Organization     types.String `tfsdk:"organization"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	Version          types.Int64  `tfsdk:"version"`
	ConnectionID     types.String `tfsdk:"connection_id"`
	Configuration    types.Map    `tfsdk:"configuration"`
	Fields           types.Set    `tfsdk:"fields"`
	AdditionalFields types.Set    `tfsdk:"additional_fields"`
	Relations        types.Set    `tfsdk:"relations"`
	Identifier       types.String `tfsdk:"identifier"`
	TrackingColumns  types.Set    `tfsdk:"tracking_columns"`
}

type modelResource struct {
	client *polytomic.Client
}

func (r *modelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data modelResourceResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var requestFields []string
	diags = data.Fields.ElementsAs(ctx, &requestFields, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var additionalRequestFields []polytomic.ModelFieldRequest
	diags = data.AdditionalFields.ElementsAs(ctx, &additionalRequestFields, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var confRequest map[string]string
	diags = data.Configuration.ElementsAs(ctx, &confRequest, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	confRequestTyped := make(map[string]interface{})
	for k, v := range confRequest {
		confRequestTyped[k] = v
	}

	var relationsRequest []polytomic.Relation
	diags = data.Relations.ElementsAs(ctx, &relationsRequest, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var trackingColumnsRequest []string
	diags = data.TrackingColumns.ElementsAs(ctx, &trackingColumnsRequest, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	request := polytomic.ModelRequest{
		Name:             data.Name.ValueString(),
		OrganizationID:   data.Organization.ValueString(),
		ConnectionID:     data.ConnectionID.ValueString(),
		Configuration:    confRequestTyped,
		Fields:           requestFields,
		AdditionalFields: additionalRequestFields,
		Relations:        relationsRequest,
		Identifier:       data.Identifier.ValueString(),
		TrackingColumns:  trackingColumnsRequest,
	}

	model, err := r.client.Models().Create(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Error creating model", err.Error())
		return
	}

	// Remove any non-string values from the configuration
	// this is a limitation of variable-typed map values seemingly not being supported
	// by the tfsdk
	for k, val := range model.Configuration {
		switch val.(type) {
		case string:
		default:
			delete(model.Configuration, k)
		}
		if val == nil || val == "" {
			delete(model.Configuration, k)
		}
	}

	config, diags := types.MapValueFrom(ctx, types.StringType, model.Configuration)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var modelFields []string
	var modelAdditionalFields []polytomic.ModelFieldRequest
	for _, field := range model.Fields {
		if !field.UserAdded {
			modelFields = append(modelFields, field.Name)
		} else {
			modelAdditionalFields = append(modelAdditionalFields, polytomic.ModelFieldRequest{
				Name:  field.Name,
				Type:  field.Type,
				Label: field.Label,
			})
		}
	}

	fields, diags := types.SetValueFrom(ctx, types.StringType, modelFields)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	additionalFields, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":  types.StringType,
			"type":  types.StringType,
			"label": types.StringType,
		},
	}, modelAdditionalFields)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if additionalFields.IsNull() {
		additionalFields, diags = types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"type":  types.StringType,
				"label": types.StringType,
			},
		}, nil)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	relations, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"to": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"model_id": types.StringType,
					"field":    types.StringType,
				},
			},
			"from": types.StringType,
		},
	}, model.Relations)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if relations.IsNull() {
		relations, diags = types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"to": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"model_id": types.StringType,
						"field":    types.StringType,
					},
				},
				"from": types.StringType,
			}}, nil)

		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	trackingColumns, diags := types.SetValueFrom(ctx, types.StringType, model.TrackingColumns)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if trackingColumns.IsNull() {
		trackingColumns, diags = types.SetValue(types.StringType, nil)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	data.ID = types.StringValue(model.ID)
	data.Organization = types.StringValue(model.OrganizationID)
	data.Name = types.StringValue(model.Name)
	data.Type = types.StringValue(model.Type)
	data.Version = types.Int64Value(int64(model.Version))
	data.ConnectionID = types.StringValue(model.ConnectionID)
	data.Configuration = config
	data.Fields = fields
	data.Relations = relations
	data.Identifier = types.StringValue(model.Identifier)
	data.TrackingColumns = trackingColumns
	data.AdditionalFields = additionalFields

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *modelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data modelResourceResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	model, err := r.client.Models().Get(ctx, data.ID.ValueString())
	if err != nil {
		pErr := polytomic.ApiError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError("Error reading model", err.Error())
		return
	}

	// Remove any non-string values from the configuration
	// this is a limitation of variable-typed map values seemingly not being supported
	// by the tfsdk
	for k, val := range model.Configuration {
		switch val.(type) {
		case string:
		default:
			delete(model.Configuration, k)
		}
		if val == nil || val == "" {
			delete(model.Configuration, k)
		}
	}

	config, diags := types.MapValueFrom(ctx, types.StringType, model.Configuration)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var modelFields []string
	var modelAdditionalFields []polytomic.ModelFieldRequest
	for _, field := range model.Fields {
		if !field.UserAdded {
			modelFields = append(modelFields, field.Name)
		} else {
			modelAdditionalFields = append(modelAdditionalFields, polytomic.ModelFieldRequest{
				Name:  field.Name,
				Type:  field.Type,
				Label: field.Label,
			})
		}
	}

	fields, diags := types.SetValueFrom(ctx, types.StringType, modelFields)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	additionalFields, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":  types.StringType,
			"type":  types.StringType,
			"label": types.StringType,
		},
	}, modelAdditionalFields)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if additionalFields.IsNull() {
		additionalFields, diags = types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"type":  types.StringType,
				"label": types.StringType,
			},
		}, nil)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	relations, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"to": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"model_id": types.StringType,
					"field":    types.StringType,
				},
			},
			"from": types.StringType,
		},
	}, model.Relations)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if relations.IsNull() {
		relations, diags = types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"to": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"model_id": types.StringType,
						"field":    types.StringType,
					},
				},
				"from": types.StringType,
			}}, nil)

		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	trackingColumns, diags := types.SetValueFrom(ctx, types.StringType, model.TrackingColumns)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if trackingColumns.IsNull() {
		trackingColumns, diags = types.SetValue(types.StringType, nil)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	data.ID = types.StringValue(model.ID)
	data.Organization = types.StringValue(model.OrganizationID)
	data.Name = types.StringValue(model.Name)
	data.Type = types.StringValue(model.Type)
	data.Version = types.Int64Value(int64(model.Version))
	data.ConnectionID = types.StringValue(model.ConnectionID)
	data.Configuration = config
	data.Fields = fields
	data.Relations = relations
	data.Identifier = types.StringValue(model.Identifier)
	data.TrackingColumns = trackingColumns
	data.AdditionalFields = additionalFields

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

func (r *modelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data modelResourceResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var requestFields []string
	diags = data.Fields.ElementsAs(ctx, &requestFields, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var additionalRequestFields []polytomic.ModelFieldRequest
	diags = data.AdditionalFields.ElementsAs(ctx, &additionalRequestFields, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var confRequest map[string]string
	diags = data.Configuration.ElementsAs(ctx, &confRequest, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	confRequestTyped := make(map[string]interface{})
	for k, v := range confRequest {
		confRequestTyped[k] = v
	}

	var relationsRequest []polytomic.Relation
	if !data.Relations.IsNull() {
		diags = data.Relations.ElementsAs(ctx, &relationsRequest, true)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	var trackingRequest []string
	if !data.TrackingColumns.IsNull() {
		diags = data.TrackingColumns.ElementsAs(ctx, &trackingRequest, true)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	model, err := r.client.Models().Update(ctx, data.ID.ValueString(),
		polytomic.ModelRequest{
			Name:             data.Name.ValueString(),
			OrganizationID:   data.Organization.ValueString(),
			ConnectionID:     data.ConnectionID.ValueString(),
			Configuration:    confRequestTyped,
			Fields:           requestFields,
			AdditionalFields: additionalRequestFields,
			Relations:        relationsRequest,
			Identifier:       data.Identifier.ValueString(),
			TrackingColumns:  trackingRequest,
		})

	if err != nil {
		resp.Diagnostics.AddError("Error updating model", err.Error())
		return
	}

	// Remove any non-string values from the configuration
	// this is a limitation of variable-typed map values seemingly not being supported
	// by the tfsdk
	for k, val := range model.Configuration {
		switch val.(type) {
		case string:
		default:
			delete(model.Configuration, k)
		}
		if val == nil || val == "" {
			delete(model.Configuration, k)
		}
	}

	config, diags := types.MapValueFrom(ctx, types.StringType, model.Configuration)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var modelFields []string
	var modelAdditionalFields []polytomic.ModelFieldRequest
	for _, field := range model.Fields {
		if !field.UserAdded {
			modelFields = append(modelFields, field.Name)
		} else {
			modelAdditionalFields = append(modelAdditionalFields, polytomic.ModelFieldRequest{
				Name:  field.Name,
				Type:  field.Type,
				Label: field.Label,
			})
		}
	}

	fields, diags := types.SetValueFrom(ctx, types.StringType, modelFields)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	additionalFields, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":  types.StringType,
			"type":  types.StringType,
			"label": types.StringType,
		},
	}, modelAdditionalFields)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if additionalFields.IsNull() {
		additionalFields, diags = types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"type":  types.StringType,
				"label": types.StringType,
			},
		}, nil)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	relations, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"to": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"model_id": types.StringType,
					"field":    types.StringType,
				},
			},
			"from": types.StringType,
		},
	}, model.Relations)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if relations.IsNull() {
		relations, diags = types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"to": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"model_id": types.StringType,
						"field":    types.StringType,
					},
				},
				"from": types.StringType,
			}}, nil)

		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	trackingColumns, diags := types.SetValueFrom(ctx, types.StringType, model.TrackingColumns)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if trackingColumns.IsNull() {
		trackingColumns, diags = types.SetValue(types.StringType, nil)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	data.ID = types.StringValue(model.ID)
	data.Organization = types.StringValue(model.OrganizationID)
	data.Name = types.StringValue(model.Name)
	data.Type = types.StringValue(model.Type)
	data.Version = types.Int64Value(int64(model.Version))
	data.ConnectionID = types.StringValue(model.ConnectionID)
	data.Configuration = config
	data.Fields = fields
	data.Relations = relations
	data.Identifier = types.StringValue(model.Identifier)
	data.TrackingColumns = trackingColumns
	data.AdditionalFields = additionalFields

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

func (r *modelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data modelResourceResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Models().Delete(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting model", err.Error())
	}

}

func (r *modelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

}
