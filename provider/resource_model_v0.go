package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type modelResourceResourceDataV0 struct {
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

var v0ModelSchema = &schema.Schema{
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
			Required:            true,
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
