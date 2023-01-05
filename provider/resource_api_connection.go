package provider

import (
	"context"
	"fmt"

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
var _ resource.Resource = &APIConnectionResource{}
var _ resource.ResourceWithImportState = &APIConnectionResource{}

func (t *APIConnectionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: ":meta:subcategory:Connection: API Connection",
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
				Required: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"url": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
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

					"body": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
					},
					"query_parameters": {
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
					"healthcheck": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Optional:            true,
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
				},
				)},
			"id": {
				Computed:            true,
				MarkdownDescription: "API Connection identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
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
	diags = data.Configuration.Attributes()["headers"].(types.Set).ElementsAs(ctx, &headers, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var params []polytomic.RequestParameter
	diags = data.Configuration.Attributes()["query_string_parameters"].(types.Set).ElementsAs(ctx, &params, true)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
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
			Type:           polytomic.APIConnectionType,
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.APIConnectionConfiguration{
				URL:                   data.Configuration.Attributes()["url"].(types.String).ValueString(),
				Headers:               headers,
				Body:                  data.Configuration.Attributes()["body"].(types.String).ValueString(),
				QueryStringParameters: params,
				Healthcheck:           data.Configuration.Attributes()["healthcheck"].(types.String).ValueString(),
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
		if err.Error() == ConnectionNotFoundErr {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading connection: %s", err))
		return
	}

	data.Id = types.StringValue(connection.ID)
	data.Name = types.StringValue(connection.Name)

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

	updated, err := r.client.Connections().Update(ctx,
		uuid.MustParse(data.Id.ValueString()),
		polytomic.UpdateConnectionMutation{
			Name:           data.Name.ValueString(),
			OrganizationId: data.Organization.ValueString(),
			Configuration:  polytomic.APIConnectionConfiguration{},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	data.Id = types.StringValue(updated.ID)
	data.Name = types.StringValue(updated.Name)

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

	err := r.client.Connections().Delete(ctx, uuid.MustParse(data.Id.ValueString()))
	if err != nil {
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
