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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &CSVConnectionResource{}
var _ resource.ResourceWithImportState = &CSVConnectionResource{}

func (t *CSVConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: CSV Connection",
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
					"url": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Sensitive:           false,
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
											}},
										Optional: true,
									},
								},
								Optional: true,
							}},
						Optional: true,
					},
				},
				Required: true,
			},
			"force_destroy": schema.BoolAttribute{
				MarkdownDescription: forceDestroyMessage,
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "CSV Connection identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *CSVConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_csv_connection"
}

type CSVConnectionResource struct {
	client *polytomic.Client
}

func (r *CSVConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
			Type:           polytomic.CsvConnectionType,
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.CSVConnectionConfiguration{
				URL:                   data.Configuration.Attributes()["url"].(types.String).ValueString(),
				Headers:               headers,
				QueryStringParameters: params,
				Auth:                  auth,
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}
	data.Id = types.StringValue(created.ID)
	data.Name = types.StringValue(created.Name)
	data.Organization = types.StringValue(created.OrganizationId)

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "CSV", "id": created.ID})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *CSVConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	data.Id = types.StringValue(connection.ID)
	data.Name = types.StringValue(connection.Name)
	data.Organization = types.StringValue(connection.OrganizationId)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *CSVConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
			Configuration: polytomic.CSVConnectionConfiguration{
				URL:                   data.Configuration.Attributes()["url"].(types.String).ValueString(),
				Headers:               headers,
				QueryStringParameters: params,
				Auth:                  auth,
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	data.Id = types.StringValue(updated.ID)
	data.Name = types.StringValue(updated.Name)
	data.Organization = types.StringValue(updated.OrganizationId)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *CSVConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *CSVConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *CSVConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
