package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &CSVConnectionResource{}
var _ resource.ResourceWithImportState = &CSVConnectionResource{}

func (t *CSVConnectionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: CSV Connection",
		Attributes: map[string]tfsdk.Attribute{
			"organization": {
				MarkdownDescription: "Organization ID",
				Optional:            true,
				Computed:            true,
				Type:                types.StringType,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"configuration": {
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"url": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"headers": {
						MarkdownDescription: "",
						Type: types.SetType{ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"name":  types.StringType,
								"value": types.StringType,
							},
						}},
						Optional: true,
					},
					"parameters": {
						MarkdownDescription: "",
						Type: types.SetType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"name":  types.StringType,
									"value": types.StringType,
								}},
						},
						Optional: true,
					},
					"auth": {
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"basic": {
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"username": {
										Type:     types.StringType,
										Optional: true,
									},
									"password": {
										Type:      types.StringType,
										Optional:  true,
										Sensitive: true,
									},
								}),
								Optional: true,
							},
							"header": {
								Type: types.ObjectType{
									AttrTypes: map[string]attr.Type{
										"name":  types.StringType,
										"value": types.StringType,
									}},
								Optional: true,
							},
							"oauth": {
								Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
									"client_id": {
										Type:     types.StringType,
										Optional: true,
									},
									"client_secret": {
										Type:      types.StringType,
										Optional:  true,
										Sensitive: true,
									},
									"token_endpoint": {
										Type:     types.StringType,
										Optional: true,
									},
									"extra_form_data": {
										Type: types.SetType{
											ElemType: types.ObjectType{
												AttrTypes: map[string]attr.Type{
													"name":  types.StringType,
													"value": types.StringType,
												}},
										},
										Optional: true,
									},
								}),
								Optional: true,
							}}),
						Optional: true,
					},
				}),
				Required: true,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Google Cloud Storage Connection identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
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
	diags = data.Configuration.Attributes()["auth"].(types.Object).As(ctx, &auth, types.ObjectAsOptions{UnhandledNullAsEmpty: true})
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
	diags = data.Configuration.Attributes()["auth"].(types.Object).As(ctx, &auth, types.ObjectAsOptions{UnhandledNullAsEmpty: true})
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

	err := r.client.Connections().Delete(ctx, uuid.MustParse(data.Id.ValueString()))
	if err != nil {
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
