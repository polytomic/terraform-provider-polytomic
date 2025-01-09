// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package connections

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
	ptcore "github.com/polytomic/polytomic-go/core"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/providerclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &CsvConnectionResource{}
var _ resource.ResourceWithImportState = &CsvConnectionResource{}

func (t *CsvConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: CSV URL Connection",
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
				Attributes: map[string]schema.Attribute{
					"auth": schema.SingleNestedAttribute{
						MarkdownDescription: `Authentication method`,
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Attributes: map[string]schema.Attribute{
							"basic": schema.SingleNestedAttribute{
								MarkdownDescription: `Basic authentication`,
								Required:            false,
								Optional:            true,
								Computed:            true,
								Sensitive:           false,
								Attributes: map[string]schema.Attribute{
									"password": schema.StringAttribute{
										MarkdownDescription: ``,
										Required:            false,
										Optional:            true,
										Computed:            true,
										Sensitive:           false,
									},
									"username": schema.StringAttribute{
										MarkdownDescription: ``,
										Required:            false,
										Optional:            true,
										Computed:            true,
										Sensitive:           false,
									},
								},
							},
							"header": schema.SingleNestedAttribute{
								MarkdownDescription: `Header key`,
								Required:            false,
								Optional:            true,
								Computed:            true,
								Sensitive:           false,
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: ``,
										Required:            false,
										Optional:            true,
										Computed:            true,
										Sensitive:           false,
									},
									"value": schema.StringAttribute{
										MarkdownDescription: ``,
										Required:            false,
										Optional:            true,
										Computed:            true,
										Sensitive:           false,
									},
								},
							},
							"oauth": schema.SingleNestedAttribute{
								MarkdownDescription: ``,
								Required:            false,
								Optional:            true,
								Computed:            true,
								Sensitive:           false,
								Attributes: map[string]schema.Attribute{
									"auth_style": schema.Int64Attribute{
										MarkdownDescription: `Auth style`,
										Required:            false,
										Optional:            true,
										Computed:            true,
										Sensitive:           false,
									},
									"client_id": schema.StringAttribute{
										MarkdownDescription: `Client ID`,
										Required:            false,
										Optional:            true,
										Computed:            true,
										Sensitive:           false,
									},
									"client_secret": schema.StringAttribute{
										MarkdownDescription: `Client secret`,
										Required:            false,
										Optional:            true,
										Computed:            true,
										Sensitive:           false,
									},
									"extra_form_data": schema.SetNestedAttribute{
										MarkdownDescription: `Extra form data`,
										Required:            false,
										Optional:            true,
										Computed:            true,
										Sensitive:           false,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													MarkdownDescription: ``,
													Required:            false,
													Optional:            true,
													Computed:            true,
													Sensitive:           false,
												},
												"value": schema.StringAttribute{
													MarkdownDescription: ``,
													Required:            false,
													Optional:            true,
													Computed:            true,
													Sensitive:           false,
												},
											},
										},
									},
									"scopes": schema.SetAttribute{
										MarkdownDescription: ``,
										Required:            false,
										Optional:            true,
										Computed:            true,
										Sensitive:           false,

										ElementType: types.StringType,
									},
									"token_endpoint": schema.StringAttribute{
										MarkdownDescription: `Token endpoint`,
										Required:            false,
										Optional:            true,
										Computed:            true,
										Sensitive:           false,
									},
								},
							},
						},
					},
					"headers": schema.SetNestedAttribute{
						MarkdownDescription: ``,
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: ``,
									Required:            false,
									Optional:            true,
									Computed:            true,
									Sensitive:           false,
								},
								"value": schema.StringAttribute{
									MarkdownDescription: ``,
									Required:            false,
									Optional:            true,
									Computed:            true,
									Sensitive:           false,
								},
							},
						},
					},
					"parameters": schema.SetNestedAttribute{
						MarkdownDescription: `Query string parameters`,
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: ``,
									Required:            false,
									Optional:            true,
									Computed:            true,
									Sensitive:           false,
								},
								"value": schema.StringAttribute{
									MarkdownDescription: ``,
									Required:            false,
									Optional:            true,
									Computed:            true,
									Sensitive:           false,
								},
							},
						},
					},
					"url": schema.StringAttribute{
						MarkdownDescription: `Base URL

    e.g. http://www.example.com`,
						Required:  true,
						Optional:  false,
						Computed:  false,
						Sensitive: false,
					},
				},

				Required: true,

				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"force_destroy": schema.BoolAttribute{
				MarkdownDescription: forceDestroyMessage,
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "CSV URL Connection identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

type CsvConf struct {
	Auth struct {
		Basic struct {
			Password string `mapstructure:"password" tfsdk:"password"`
			Username string `mapstructure:"username" tfsdk:"username"`
		} `mapstructure:"basic" tfsdk:"basic"`
		Header struct {
			Name  string `mapstructure:"name" tfsdk:"name"`
			Value string `mapstructure:"value" tfsdk:"value"`
		} `mapstructure:"header" tfsdk:"header"`
		Oauth struct {
			Auth_style      int64  `mapstructure:"auth_style" tfsdk:"auth_style"`
			Client_id       string `mapstructure:"client_id" tfsdk:"client_id"`
			Client_secret   string `mapstructure:"client_secret" tfsdk:"client_secret"`
			Extra_form_data []struct {
				Name  string `mapstructure:"name" tfsdk:"name"`
				Value string `mapstructure:"value" tfsdk:"value"`
			} `mapstructure:"extra_form_data" tfsdk:"extra_form_data"`
			Scopes         []string `mapstructure:"scopes" tfsdk:"scopes"`
			Token_endpoint string   `mapstructure:"token_endpoint" tfsdk:"token_endpoint"`
		} `mapstructure:"oauth" tfsdk:"oauth"`
	} `mapstructure:"auth" tfsdk:"auth"`
	Headers []struct {
		Name  string `mapstructure:"name" tfsdk:"name"`
		Value string `mapstructure:"value" tfsdk:"value"`
	} `mapstructure:"headers" tfsdk:"headers"`
	Parameters []struct {
		Name  string `mapstructure:"name" tfsdk:"name"`
		Value string `mapstructure:"value" tfsdk:"value"`
	} `mapstructure:"parameters" tfsdk:"parameters"`
	Url string `mapstructure:"url" tfsdk:"url"`
}

type CsvConnectionResource struct {
	provider *providerclient.Provider
}

func (r *CsvConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *CsvConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_csv_connection"
}

func (r *CsvConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	connConf, err := objectMapValue(ctx, data.Configuration)
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection configuration", err.Error())
		return
	}
	created, err := client.Connections.Create(ctx, &polytomic.CreateConnectionRequestSchema{
		Name:           data.Name.ValueString(),
		Type:           "csv",
		OrganizationId: data.Organization.ValueStringPointer(),
		Configuration:  connConf,
		Validate:       pointer.ToBool(false),
	})
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error creating connection: %s", err))
		return
	}
	data.Id = types.StringPointerValue(created.Data.Id)
	data.Name = types.StringPointerValue(created.Data.Name)
	data.Organization = types.StringPointerValue(created.Data.OrganizationId)

	conf := CsvConf{}
	err = mapstructure.Decode(created.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"auth": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"basic": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"password": types.StringType,
						"username": types.StringType,
					},
				}, "header": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"value": types.StringType,
					},
				}, "oauth": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"auth_style":    types.NumberType,
						"client_id":     types.StringType,
						"client_secret": types.StringType,
						"extra_form_data": types.SetType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"name":  types.StringType,
									"value": types.StringType,
								},
							},
						},
						"scopes": types.SetType{
							ElemType: types.StringType,
						},
						"token_endpoint": types.StringType,
					},
				},
			},
		}, "headers": types.SetType{
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
		"url": types.StringType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Csv", "id": created.Data.Id})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *CsvConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionData

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
	connection, err := client.Connections.Get(ctx, data.Id.ValueString())
	if err != nil {
		pErr := &ptcore.APIError{}
		if errors.As(err, &pErr) {
			if pErr.StatusCode == http.StatusNotFound {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error reading connection: %s", err))
		return
	}
	data.Id = types.StringPointerValue(connection.Data.Id)
	data.Name = types.StringPointerValue(connection.Data.Name)
	data.Organization = types.StringPointerValue(connection.Data.OrganizationId)

	conf := CsvConf{}
	err = mapstructure.Decode(connection.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"auth": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"basic": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"password": types.StringType,
						"username": types.StringType,
					},
				}, "header": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"value": types.StringType,
					},
				}, "oauth": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"auth_style":    types.NumberType,
						"client_id":     types.StringType,
						"client_secret": types.StringType,
						"extra_form_data": types.SetType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"name":  types.StringType,
									"value": types.StringType,
								},
							},
						},
						"scopes": types.SetType{
							ElemType: types.StringType,
						},
						"token_endpoint": types.StringType,
					},
				},
			},
		}, "headers": types.SetType{
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
		"url": types.StringType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *CsvConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data connectionData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	connConf, err := objectMapValue(ctx, data.Configuration)
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection configuration", err.Error())
		return
	}
	updated, err := client.Connections.Update(ctx,
		data.Id.ValueString(),
		&polytomic.UpdateConnectionRequestSchema{
			Name:           data.Name.ValueString(),
			OrganizationId: data.Organization.ValueStringPointer(),
			Configuration:  connConf,
			Validate:       pointer.ToBool(false),
		})
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	data.Id = types.StringPointerValue(updated.Data.Id)
	data.Name = types.StringPointerValue(updated.Data.Name)
	data.Organization = types.StringPointerValue(updated.Data.OrganizationId)

	conf := CsvConf{}
	err = mapstructure.Decode(updated.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"auth": types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"basic": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"password": types.StringType,
						"username": types.StringType,
					},
				}, "header": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"value": types.StringType,
					},
				}, "oauth": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"auth_style":    types.NumberType,
						"client_id":     types.StringType,
						"client_secret": types.StringType,
						"extra_form_data": types.SetType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"name":  types.StringType,
									"value": types.StringType,
								},
							},
						},
						"scopes": types.SetType{
							ElemType: types.StringType,
						},
						"token_endpoint": types.StringType,
					},
				},
			},
		}, "headers": types.SetType{
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
		"url": types.StringType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *CsvConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionData

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
	if data.ForceDestroy.ValueBool() {
		err := client.Connections.Remove(ctx, data.Id.ValueString(), &polytomic.ConnectionsRemoveRequest{
			Force: pointer.ToBool(true),
		})
		if err != nil {
			pErr := &polytomic.NotFoundError{}
			if errors.As(err, &pErr) {
				resp.State.RemoveResource(ctx)
				return
			}

			resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error deleting connection: %s", err))
		}
		return
	}

	err = client.Connections.Remove(ctx, data.Id.ValueString(), &polytomic.ConnectionsRemoveRequest{
		Force: pointer.ToBool(false),
	})
	if err != nil {
		pErr := &polytomic.NotFoundError{}
		if errors.As(err, &pErr) {
			resp.State.RemoveResource(ctx)
			return
		}
	}
	pErr := &polytomic.UnprocessableEntityError{}
	if errors.As(err, &pErr) {
		if strings.Contains(*pErr.Body.Message, "connection in use") {
			if used_by, ok := pErr.Body.Metadata["used_by"].([]interface{}); ok {
				for _, us := range used_by {
					if user, ok := us.(map[string]interface{}); ok {
						resp.Diagnostics.AddError("Connection in use",
							fmt.Sprintf("Connection is used by %s \"%s\" (%s). Please remove before deleting this connection.",
								user["type"], user["name"], user["id"]),
						)
					}
				}
				return
			}
		}
	}

	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error deleting connection: %s", err))
	}
}

func (r *CsvConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
