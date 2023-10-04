package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
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
	client *polytomic.Client
}

func (r *APIConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var headers []polytomic.RequestParameter
	if data.Configuration.Attributes()["headers"] != nil {
		diags = data.Configuration.Attributes()["headers"].(types.Set).ElementsAs(ctx, &headers, true)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	var params []polytomic.RequestParameter
	if data.Configuration.Attributes()["parameters"] != nil {
		diags = data.Configuration.Attributes()["parameters"].(types.Set).ElementsAs(ctx, &params, true)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	var auth polytomic.Auth
	diags = data.Configuration.Attributes()["auth"].(types.Object).As(ctx, &auth, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	created, err := r.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:           data.Name.ValueString(),
			Type:           polytomic.APIConnectionType,
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.APIConnectionConfiguration{
				URL:         data.Configuration.Attributes()["url"].(types.String).ValueString(),
				Headers:     headers,
				Parameters:  params,
				Healthcheck: data.Configuration.Attributes()["healthcheck"].(types.String).ValueString(),
				Auth:        auth,
			},
		},
		polytomic.WithIdempotencyKey(uuid.NewString()),
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}

	var output polytomic.APIConnectionConfiguration
	cfg := &mapstructure.DecoderConfig{
		Result: &output,
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	decoder.Decode(created.Configuration)
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
	}, output)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringValue(created.ID)
	data.Name = types.StringValue(created.Name)
	data.Organization = types.StringValue(created.OrganizationId)

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "API", "id": created.ID})

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

	connection, err := r.client.Connections().Get(ctx, uuid.MustParse(data.Id.ValueString()))
	if err != nil {
		pErr := polytomic.ApiError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
			if strings.Contains(pErr.Message, "connection in use") {
				for _, meta := range pErr.Metadata {
					info := meta.(map[string]interface{})
					resp.Diagnostics.AddError("Connection in use",
						fmt.Sprintf("Connection is used by %s \"%s\" (%s). Please remove before deleting this connection.",
							info["type"], info["name"], info["id"]),
					)
				}
				return
			}
		}
	}

	var output polytomic.APIConnectionConfiguration
	cfg := &mapstructure.DecoderConfig{
		Result: &output,
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	decoder.Decode(connection.Configuration)
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
	}, output)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringValue(connection.ID)
	data.Name = types.StringValue(connection.Name)
	data.Organization = types.StringValue(connection.OrganizationId)

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

	var headers []polytomic.RequestParameter
	if data.Configuration.Attributes()["headers"] != nil {
		diags = data.Configuration.Attributes()["headers"].(types.Set).ElementsAs(ctx, &headers, true)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	var params []polytomic.RequestParameter
	if data.Configuration.Attributes()["parameters"] != nil {
		diags = data.Configuration.Attributes()["parameters"].(types.Set).ElementsAs(ctx, &params, true)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	var auth polytomic.Auth
	diags = data.Configuration.Attributes()["auth"].(types.Object).As(ctx, &auth, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	updated, err := r.client.Connections().Update(ctx,
		uuid.MustParse(data.Id.ValueString()),
		polytomic.UpdateConnectionMutation{
			Name:           data.Name.ValueString(),
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.APIConnectionConfiguration{
				URL:         data.Configuration.Attributes()["url"].(types.String).ValueString(),
				Headers:     headers,
				Parameters:  params,
				Healthcheck: data.Configuration.Attributes()["healthcheck"].(types.String).ValueString(),
				Auth:        auth,
			},
		},
		polytomic.WithIdempotencyKey(uuid.NewString()),
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	var output polytomic.APIConnectionConfiguration
	cfg := &mapstructure.DecoderConfig{
		Result: &output,
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	decoder.Decode(updated.Configuration)
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
	}, output)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Id = types.StringValue(updated.ID)
	data.Name = types.StringValue(updated.Name)
	data.Organization = types.StringValue(updated.OrganizationId)

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
		err := r.client.Connections().Delete(ctx, uuid.MustParse(data.Id.ValueString()), polytomic.WithForceDelete())
		if err != nil {
			pErr := polytomic.ApiError{}
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

	err := r.client.Connections().Delete(ctx, uuid.MustParse(data.Id.ValueString()))
	if err != nil {
		pErr := polytomic.ApiError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
			if strings.Contains(pErr.Message, "connection in use") {
				for _, meta := range pErr.Metadata {
					info := meta.(map[string]interface{})
					resp.Diagnostics.AddError("Connection in use",
						fmt.Sprintf("Connection is used by %s \"%s\" (%s). Please remove before deleting this connection.",
							info["type"], info["name"], info["id"]),
					)
				}
				return
			}
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
