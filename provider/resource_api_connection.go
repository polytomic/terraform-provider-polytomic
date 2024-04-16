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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
	ptclient "github.com/polytomic/polytomic-go/client"
	ptcore "github.com/polytomic/polytomic-go/core"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &APIConnectionResource{}
var _ resource.ResourceWithImportState = &APIConnectionResource{}

func (t *APIConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: API Connection",
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"configuration": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
					},
					"headers": schema.SetAttribute{
						MarkdownDescription: "",
						ElementType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name":  types.StringType,
								"value": types.StringType,
							},
						},
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
					},
					"parameters": schema.SetAttribute{
						MarkdownDescription: "",
						ElementType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name":  types.StringType,
								"value": types.StringType,
							},
						},
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
					},
					"healthcheck": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
					},
					"body": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
					"auth": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"basic": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"username": schema.StringAttribute{
										Optional: true,
									},
									"password": schema.StringAttribute{
										Optional:  true,
										Sensitive: true,
									},
								},
								Optional: true,
							},
							"header": schema.ObjectAttribute{
								AttributeTypes: map[string]attr.Type{
									"name":  types.StringType,
									"value": types.StringType,
								},
								Optional: true,
							},
							"oauth": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"client_id": schema.StringAttribute{
										Optional: true,
									},
									"client_secret": schema.StringAttribute{
										Optional:  true,
										Sensitive: true,
									},
									"token_endpoint": schema.StringAttribute{
										Optional: true,
									},
									"extra_form_data": schema.SetAttribute{
										ElementType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"name":  types.StringType,
												"value": types.StringType,
											},
										},
										Optional: true,
									},
								},
								Optional: true,
							}},
						Optional: true,
					},
				},
			},
			"force_destroy": schema.BoolAttribute{
				MarkdownDescription: forceDestroyMessage,
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "API Connection identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *APIConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_connection"
}

type APIConnectionResource struct {
	client *ptclient.Client
}

type APIConf struct {
	URL         string             `json:"url" mapstructure:"url" tfsdk:"url"`
	Headers     []RequestParameter `json:"headers" mapstructure:"headers" tfsdk:"headers"`
	Body        string             `json:"body" mapstructure:"body" tfsdk:"body"`
	Parameters  []RequestParameter `json:"parameters" mapstructure:"parameters" tfsdk:"parameters"`
	Healthcheck string             `json:"healthcheck" mapstructure:"healthcheck" tfsdk:"healthcheck"`
	Auth        Auth               `json:"auth" mapstructure:",squash" tfsdk:"auth"`
}

func (r *APIConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var headers []RequestParameter
	if data.Configuration.Attributes()["headers"] != nil {
		diags = data.Configuration.Attributes()["headers"].(types.Set).ElementsAs(ctx, &headers, true)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	var params []RequestParameter
	if data.Configuration.Attributes()["parameters"] != nil {
		diags = data.Configuration.Attributes()["parameters"].(types.Set).ElementsAs(ctx, &params, true)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	var auth Auth
	diags = data.Configuration.Attributes()["auth"].(types.Object).As(ctx, &auth, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	created, err := r.client.Connections.Create(ctx,
		&polytomic.CreateConnectionRequestSchema{
			Name:           data.Name.ValueString(),
			Type:           "api",
			OrganizationId: data.Organization.ValueStringPointer(),
			Configuration: map[string]interface{}{
				"url":         data.Configuration.Attributes()["url"].(types.String).ValueString(),
				"headers":     headers,
				"parameters":  params,
				"healthcheck": data.Configuration.Attributes()["healthcheck"].(types.String).ValueString(),
				"auth":        auth,
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}

	conf := APIConf{}
	err = mapstructure.Decode(created.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"url": types.StringType,
		"headers": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				},
			},
		},
		"parameters": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				},
			},
		},
		"body":        types.StringType,
		"healthcheck": types.StringType,
		"auth": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"basic": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"username": types.StringType,
						"password": types.StringType,
					},
				},
				"header": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"value": types.StringType,
					},
				},
				"oauth": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"client_id":      types.StringType,
						"client_secret":  types.StringType,
						"token_endpoint": types.StringType,
						"extra_form_data": types.SetType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"name":  types.StringType,
									"value": types.StringType,
								},
							},
						},
					},
				},
			},
		},
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringPointerValue(created.Data.Id)
	data.Name = types.StringPointerValue(created.Data.Name)
	data.Organization = types.StringPointerValue(created.Data.OrganizationId)

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "API", "id": created.Data.Id})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *APIConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	connection, err := r.client.Connections.Get(ctx, data.Id.ValueString())
	if err != nil {
		pErr := &ptcore.APIError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
			// if strings.Contains(pErr.Message, "connection in use") {
			// 	for _, meta := range pErr.Metadata {
			// 		info := meta.(map[string]interface{})
			// 		resp.Diagnostics.AddError("Connection in use",
			// 			fmt.Sprintf("Connection is used by %s \"%s\" (%s). Please remove before deleting this connection.",
			// 				info["type"], info["name"], info["id"]),
			// 		)
			// 	}
			// 	return
			// }
		}
	}

	conf := APIConf{}
	err = mapstructure.Decode(connection.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"url": types.StringType,
		"headers": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				},
			},
		},
		"parameters": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				},
			},
		},
		"body":        types.StringType,
		"healthcheck": types.StringType,
		"auth": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"basic": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"username": types.StringType,
						"password": types.StringType,
					},
				},
				"header": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"value": types.StringType,
					},
				},
				"oauth": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"client_id":      types.StringType,
						"client_secret":  types.StringType,
						"token_endpoint": types.StringType,
						"extra_form_data": types.SetType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"name":  types.StringType,
									"value": types.StringType,
								},
							},
						},
					},
				},
			},
		},
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringPointerValue(connection.Data.Id)
	data.Name = types.StringPointerValue(connection.Data.Name)
	data.Organization = types.StringPointerValue(connection.Data.OrganizationId)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *APIConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data connectionData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var headers []RequestParameter
	if data.Configuration.Attributes()["headers"] != nil {
		diags = data.Configuration.Attributes()["headers"].(types.Set).ElementsAs(ctx, &headers, true)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	var params []RequestParameter
	if data.Configuration.Attributes()["parameters"] != nil {
		diags = data.Configuration.Attributes()["parameters"].(types.Set).ElementsAs(ctx, &params, true)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	var auth Auth
	diags = data.Configuration.Attributes()["auth"].(types.Object).As(ctx, &auth, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	updated, err := r.client.Connections.Update(ctx,
		data.Id.ValueString(),
		&polytomic.UpdateConnectionRequestSchema{
			Name:           data.Name.ValueString(),
			OrganizationId: data.Organization.ValueStringPointer(),
			Configuration: map[string]interface{}{
				"url":         data.Configuration.Attributes()["url"].(types.String).ValueString(),
				"headers":     headers,
				"parameters":  params,
				"healthcheck": data.Configuration.Attributes()["healthcheck"].(types.String).ValueString(),
				"auth":        auth,
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	conf := APIConf{}
	err = mapstructure.Decode(updated.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"url": types.StringType,
		"headers": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				},
			},
		},
		"parameters": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
				},
			},
		},
		"body":        types.StringType,
		"healthcheck": types.StringType,
		"auth": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"basic": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"username": types.StringType,
						"password": types.StringType,
					},
				},
				"header": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"value": types.StringType,
					},
				},
				"oauth": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"client_id":      types.StringType,
						"client_secret":  types.StringType,
						"token_endpoint": types.StringType,
						"extra_form_data": types.SetType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"name":  types.StringType,
									"value": types.StringType,
								},
							},
						},
					},
				},
			},
		},
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringPointerValue(updated.Data.Id)
	data.Name = types.StringPointerValue(updated.Data.Name)
	data.Organization = types.StringPointerValue(updated.Data.OrganizationId)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *APIConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ForceDestroy.ValueBool() {
		err := r.client.Connections.Remove(ctx, data.Id.ValueString(), &polytomic.ConnectionsRemoveRequest{Force: pointer.ToBool(true)})
		if err != nil {
			pErr := &ptcore.APIError{}
			if errors.As(err, &pErr) {
				if pErr.StatusCode == http.StatusNotFound {
					resp.State.RemoveResource(ctx)
					return
				}
			}
			resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting connection: %s", err))
		}
		return
	}

	err := r.client.Connections.Remove(ctx, data.Id.ValueString(), &polytomic.ConnectionsRemoveRequest{})
	if err != nil {
		pErr := &ptcore.APIError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
			// if strings.Contains(pErr.Message, "connection in use") {
			// 	for _, meta := range pErr.Metadata {
			// 		info := meta.(map[string]interface{})
			// 		resp.Diagnostics.AddError("Connection in use",
			// 			fmt.Sprintf("Connection is used by %s \"%s\" (%s). Please remove before deleting this connection.",
			// 				info["type"], info["name"], info["id"]),
			// 		)
			// 	}
			// 	return
			// }
		}

		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting connection: %s", err))
		return
	}
}

func (r *APIConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *APIConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
