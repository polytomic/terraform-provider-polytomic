package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
	ptclient "github.com/polytomic/polytomic-go/client"
	ptcore "github.com/polytomic/polytomic-go/core"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &modelResource{}
var _ resource.ResourceWithImportState = &modelResource{}

func (r *modelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Models: Model",
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
			},
			"connection_id": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"configuration": schema.MapAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				// PlanModifiers: []planmodifier.String{
				// 	stringplanmodifier.UseStateForUnknown(),
				// },
			},
			"fields": schema.SetAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"additional_fields": schema.SetAttribute{
				MarkdownDescription: "",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"type":  types.StringType,
						"label": types.StringType,
					},
				},
				Optional: true,
				Computed: true,
			},
			"relations": schema.SetAttribute{
				MarkdownDescription: "",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"to": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"model_id": types.StringType,
								"field":    types.StringType,
							},
						},
						"from": types.StringType,
					},
				},
				Optional: true,
				Computed: true,
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tracking_columns": schema.SetAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *modelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	client *ptclient.Client
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

	var additionalRequestFields []*polytomic.ModelModelFieldRequest
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

	var relationsRequest []*polytomic.ModelRelation
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

	request := &polytomic.CreateModelRequest{
		Name:             data.Name.ValueString(),
		ConnectionId:     data.ConnectionID.ValueString(),
		Configuration:    confRequestTyped,
		Fields:           requestFields,
		AdditionalFields: additionalRequestFields,
		Relations:        relationsRequest,
		TrackingColumns:  trackingColumnsRequest,
	}
	if !data.Identifier.IsNull() && data.Identifier.ValueString() != "" {
		request.Identifier = data.Identifier.ValueStringPointer()
	}
	if !data.Organization.IsNull() && data.Organization.ValueString() != "" {
		request.OrganizationId = data.Organization.ValueStringPointer()
	}

	model, err := r.client.Models.Create(ctx, &polytomic.ModelsCreateRequest{Body: request})
	if err != nil {
		resp.Diagnostics.AddError("Error creating model", err.Error())
		return
	}

	// Remove any non-string values from the configuration
	// this is a limitation of variable-typed map values seemingly not being supported
	// by the tfsdk
	for k, val := range model.Data.Configuration {
		switch val.(type) {
		case string:
		default:
			delete(model.Data.Configuration, k)
		}
		if val == nil || val == "" {
			delete(model.Data.Configuration, k)
		}
	}

	config, diags := types.MapValueFrom(ctx, types.StringType, model.Data.Configuration)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error creating model", err.Error())
		return
	}

	var modelFields []*string
	var modelAdditionalFields []polytomic.ModelModelFieldRequest
	for _, field := range model.Data.Fields {
		if !pointer.GetBool(field.UserAdded) {
			modelFields = append(modelFields, field.Name)
		} else {
			modelAdditionalFields = append(modelAdditionalFields, polytomic.ModelModelFieldRequest{
				Name:  pointer.GetString(field.Name),
				Type:  pointer.GetString(field.Type),
				Label: pointer.GetString(field.Label),
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
	}, model.Data.Relations)
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

	trackingColumns, diags := types.SetValueFrom(ctx, types.StringType, model.Data.TrackingColumns)
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

	data.ID = types.StringPointerValue(model.Data.Id)
	data.Organization = types.StringPointerValue(model.Data.OrganizationId)
	data.Name = types.StringPointerValue(model.Data.Name)
	data.Type = types.StringPointerValue(model.Data.Type)
	data.Version = types.Int64Value(int64(pointer.GetInt(model.Data.Version)))
	data.ConnectionID = types.StringPointerValue(model.Data.ConnectionId)
	data.Configuration = config
	data.Fields = fields
	data.Relations = relations
	data.Identifier = types.StringValue(pointer.Get(model.Data.Identifier))
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

	model, err := r.client.Models.Get(ctx, data.ID.ValueString(), nil)
	if err != nil {
		pErr := &ptcore.APIError{}
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
	for k, val := range model.Data.Configuration {
		switch val.(type) {
		case string:
		default:
			delete(model.Data.Configuration, k)
		}
		if val == nil || val == "" {
			delete(model.Data.Configuration, k)
		}
	}

	config, diags := types.MapValueFrom(ctx, types.StringType, model.Data.Configuration)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error creating model", err.Error())
		return
	}

	var modelFields []*string
	var modelAdditionalFields []polytomic.ModelModelFieldRequest
	for _, field := range model.Data.Fields {
		if !pointer.GetBool(field.UserAdded) {
			modelFields = append(modelFields, field.Name)
		} else {
			modelAdditionalFields = append(modelAdditionalFields, polytomic.ModelModelFieldRequest{
				Name:  pointer.GetString(field.Name),
				Type:  pointer.GetString(field.Type),
				Label: pointer.GetString(field.Label),
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
	}, model.Data.Relations)
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

	trackingColumns, diags := types.SetValueFrom(ctx, types.StringType, model.Data.TrackingColumns)
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

	data.ID = types.StringPointerValue(model.Data.Id)
	data.Organization = types.StringPointerValue(model.Data.OrganizationId)
	data.Name = types.StringPointerValue(model.Data.Name)
	data.Type = types.StringPointerValue(model.Data.Type)
	data.Version = types.Int64Value(int64(pointer.GetInt(model.Data.Version)))
	data.ConnectionID = types.StringPointerValue(model.Data.ConnectionId)
	data.Configuration = config
	data.Fields = fields
	data.Relations = relations
	data.Identifier = types.StringValue(pointer.Get(model.Data.Identifier))
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

	var additionalRequestFields []*polytomic.ModelModelFieldRequest
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

	var relationsRequest []*polytomic.ModelRelation
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

	request := &polytomic.UpdateModelRequest{
		Name:             data.Name.ValueString(),
		ConnectionId:     data.ConnectionID.ValueString(),
		Configuration:    confRequestTyped,
		Fields:           requestFields,
		AdditionalFields: additionalRequestFields,
		Relations:        relationsRequest,
		TrackingColumns:  trackingRequest,
	}
	if !data.Identifier.IsNull() && data.Identifier.ValueString() != "" {
		request.Identifier = data.Identifier.ValueStringPointer()
	}
	if !data.Organization.IsNull() && data.Organization.ValueString() != "" {
		request.OrganizationId = data.Organization.ValueStringPointer()
	}

	model, err := r.client.Models.Update(ctx, data.ID.ValueString(), request)
	if err != nil {
		resp.Diagnostics.AddError("Error updating model", err.Error())
		return
	}

	// Remove any non-string values from the configuration
	// this is a limitation of variable-typed map values seemingly not being supported
	// by the tfsdk
	for k, val := range model.Data.Configuration {
		switch val.(type) {
		case string:
		default:
			delete(model.Data.Configuration, k)
		}
		if val == nil || val == "" {
			delete(model.Data.Configuration, k)
		}
	}

	config, diags := types.MapValueFrom(ctx, types.StringType, model.Data.Configuration)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var modelFields []*string
	var modelAdditionalFields []polytomic.ModelModelFieldRequest
	for _, field := range model.Data.Fields {
		if !pointer.GetBool(field.UserAdded) {
			modelFields = append(modelFields, field.Name)
		} else {
			modelAdditionalFields = append(modelAdditionalFields, polytomic.ModelModelFieldRequest{
				Name:  pointer.GetString(field.Name),
				Type:  pointer.GetString(field.Type),
				Label: pointer.GetString(field.Label),
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
	}, model.Data.Relations)
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

	trackingColumns, diags := types.SetValueFrom(ctx, types.StringType, model.Data.TrackingColumns)
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

	data.ID = types.StringPointerValue(model.Data.Id)
	data.Organization = types.StringPointerValue(model.Data.OrganizationId)
	data.Name = types.StringPointerValue(model.Data.Name)
	data.Type = types.StringPointerValue(model.Data.Type)
	data.Version = types.Int64Value(int64(pointer.GetInt(model.Data.Version)))
	data.ConnectionID = types.StringPointerValue(model.Data.ConnectionId)
	data.Configuration = config
	data.Fields = fields
	data.Relations = relations
	data.Identifier = types.StringValue(pointer.Get(model.Data.Identifier))
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

	err := r.client.Models.Remove(ctx, data.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting model", err.Error())
	}

}

func (r *modelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

}
