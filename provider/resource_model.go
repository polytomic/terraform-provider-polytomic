package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
	ptcore "github.com/polytomic/polytomic-go/core"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &modelResource{}
var _ resource.ResourceWithImportState = &modelResource{}
var _ resource.ResourceWithUpgradeState = &modelResource{}

func (r *modelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             1,
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
			"configuration": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"fields": schema.SetAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"additional_fields": schema.SetNestedAttribute{
				MarkdownDescription: "",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "",
							Required:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: "",
							Required:            true,
						},
					},
				},
				Optional: true,
				Computed: true,
			},
			"relations": schema.SetNestedAttribute{
				MarkdownDescription: "",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"to": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"model_id": schema.StringAttribute{
									Optional: true,
								},
								"field": schema.StringAttribute{
									Optional: true,
								},
							},
							Optional: true,
						},
						"from": schema.StringAttribute{
							Optional: true,
						},
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
				Default: stringdefault.StaticString(""),
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
			"policies": schema.SetAttribute{
				MarkdownDescription: "Policy IDs attached to this model",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the model was created",
				Computed:            true,
				CustomType:          timetypes.RFC3339Type{},
			},
			"created_by": schema.SingleNestedAttribute{
				MarkdownDescription: "Actor who created this model",
				Computed:            true,
				Attributes:          actorAttributes(),
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the model was last updated",
				Computed:            true,
				CustomType:          timetypes.RFC3339Type{},
			},
			"updated_by": schema.SingleNestedAttribute{
				MarkdownDescription: "Actor who last updated this model",
				Computed:            true,
				Attributes:          actorAttributes(),
			},
		},
	}
}

type modelResourceResourceData struct {
	ID               types.String      `tfsdk:"id"`
	Organization     types.String      `tfsdk:"organization"`
	Name             types.String      `tfsdk:"name"`
	Type             types.String      `tfsdk:"type"`
	Version          types.Int64       `tfsdk:"version"`
	ConnectionID     types.String      `tfsdk:"connection_id"`
	Configuration    types.String      `tfsdk:"configuration"`
	Fields           types.Set         `tfsdk:"fields"`
	AdditionalFields types.Set         `tfsdk:"additional_fields"`
	Relations        types.Set         `tfsdk:"relations"`
	Identifier       types.String      `tfsdk:"identifier"`
	TrackingColumns  types.Set         `tfsdk:"tracking_columns"`
	Policies         types.Set         `tfsdk:"policies"`
	CreatedAt        timetypes.RFC3339 `tfsdk:"created_at"`
	CreatedBy        types.Object      `tfsdk:"created_by"`
	UpdatedAt        timetypes.RFC3339 `tfsdk:"updated_at"`
	UpdatedBy        types.Object      `tfsdk:"updated_by"`
}

type modelResource struct {
	provider *providerclient.Provider
}

func (r *modelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *modelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
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

	var confRequest map[string]interface{}
	err := json.Unmarshal([]byte(data.Configuration.ValueString()), &confRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error creating model", err.Error())
		return
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
		Configuration:    confRequest,
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

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	model, err := client.Models.Create(ctx,
		&polytomic.ModelsCreateRequest{Body: request},
	)
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

	enc, err := json.Marshal(model.Data.Configuration)
	if err != nil {
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
	data.Configuration = types.StringValue(string(enc))
	data.Fields = fields
	data.Relations = relations
	data.Identifier = types.StringValue(pointer.Get(model.Data.Identifier))
	data.TrackingColumns = trackingColumns
	data.AdditionalFields = additionalFields

	// Policies
	data.Policies, diags = types.SetValueFrom(ctx, types.StringType, model.Data.Policies)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Audit fields
	if model.Data.CreatedAt != nil {
		data.CreatedAt = timetypes.NewRFC3339TimeValue(*model.Data.CreatedAt)
	}
	if model.Data.CreatedBy != nil {
		data.CreatedBy, diags = types.ObjectValueFrom(ctx, actorAttrTypes(), model.Data.CreatedBy)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if model.Data.UpdatedAt != nil {
		data.UpdatedAt = timetypes.NewRFC3339TimeValue(*model.Data.UpdatedAt)
	}
	if model.Data.UpdatedBy != nil {
		data.UpdatedBy, diags = types.ObjectValueFrom(ctx, actorAttrTypes(), model.Data.UpdatedBy)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

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

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	model, err := client.Models.Get(ctx, data.ID.ValueString(), &polytomic.ModelsGetRequest{})
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

	enc, err := json.Marshal(model.Data.Configuration)
	if err != nil {
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
	data.Configuration = types.StringValue(string(enc))
	data.Fields = fields
	data.Relations = relations
	data.Identifier = types.StringValue(pointer.Get(model.Data.Identifier))
	data.TrackingColumns = trackingColumns
	data.AdditionalFields = additionalFields

	// Policies
	data.Policies, diags = types.SetValueFrom(ctx, types.StringType, model.Data.Policies)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Audit fields
	if model.Data.CreatedAt != nil {
		data.CreatedAt = timetypes.NewRFC3339TimeValue(*model.Data.CreatedAt)
	}
	if model.Data.CreatedBy != nil {
		data.CreatedBy, diags = types.ObjectValueFrom(ctx, actorAttrTypes(), model.Data.CreatedBy)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if model.Data.UpdatedAt != nil {
		data.UpdatedAt = timetypes.NewRFC3339TimeValue(*model.Data.UpdatedAt)
	}
	if model.Data.UpdatedBy != nil {
		data.UpdatedBy, diags = types.ObjectValueFrom(ctx, actorAttrTypes(), model.Data.UpdatedBy)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

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

	var confRequest map[string]interface{}
	err := json.Unmarshal([]byte(data.Configuration.ValueString()), &confRequest)
	if err != nil {
		resp.Diagnostics.AddError("Error creating model", err.Error())
		return
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
		Configuration:    confRequest,
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

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	model, err := client.Models.Update(ctx, data.ID.ValueString(), request)
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

	enc, err := json.Marshal(model.Data.Configuration)
	if err != nil {
		resp.Diagnostics.AddError("Error updating model", err.Error())
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
	data.Configuration = types.StringValue(string(enc))
	data.Fields = fields
	data.Relations = relations
	data.Identifier = types.StringValue(pointer.Get(model.Data.Identifier))
	data.TrackingColumns = trackingColumns
	data.AdditionalFields = additionalFields

	// Policies
	data.Policies, diags = types.SetValueFrom(ctx, types.StringType, model.Data.Policies)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Audit fields
	if model.Data.CreatedAt != nil {
		data.CreatedAt = timetypes.NewRFC3339TimeValue(*model.Data.CreatedAt)
	}
	if model.Data.CreatedBy != nil {
		data.CreatedBy, diags = types.ObjectValueFrom(ctx, actorAttrTypes(), model.Data.CreatedBy)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}
	if model.Data.UpdatedAt != nil {
		data.UpdatedAt = timetypes.NewRFC3339TimeValue(*model.Data.UpdatedAt)
	}
	if model.Data.UpdatedBy != nil {
		data.UpdatedBy, diags = types.ObjectValueFrom(ctx, actorAttrTypes(), model.Data.UpdatedBy)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

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

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	err = client.Models.Remove(ctx, data.ID.ValueString(), &polytomic.ModelsRemoveRequest{})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting model", err.Error())
		return
	}
}

func (r *modelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *modelResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 (prior state version) to 1 (Schema.Version)
		0: {
			PriorSchema: v0ModelSchema,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorStateData modelResourceResourceDataV0

				resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)

				if resp.Diagnostics.HasError() {
					return
				}
				confBytes, err := json.Marshal(priorStateData.Configuration.Elements())
				if err != nil {
					resp.Diagnostics.AddError("Error marshalling configuration", err.Error())
					return
				}

				upgradedStateData := modelResourceResourceData{
					ID:               priorStateData.ID,
					Organization:     priorStateData.Organization,
					Name:             priorStateData.Name,
					Type:             priorStateData.Type,
					Version:          priorStateData.Version,
					Configuration:    types.StringValue(string(confBytes)),
					Fields:           priorStateData.Fields,
					AdditionalFields: priorStateData.AdditionalFields,
					Relations:        priorStateData.Relations,
					Identifier:       priorStateData.Identifier,
					TrackingColumns:  priorStateData.TrackingColumns,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedStateData)...)
			},
		},
	}
}
