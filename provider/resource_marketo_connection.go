// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &MarketoConnectionResource{}
var _ resource.ResourceWithImportState = &MarketoConnectionResource{}

func (t *MarketoConnectionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: ":meta:subcategory:Connection: Marketo Connection",
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
					"client_id": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"client_secret": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           true,
					},
					"rest_endpoint": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"enforce_api_limits": {
						MarkdownDescription: "",
						Type:                types.BoolType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"daily_api_calls": {
						MarkdownDescription: "",
						Type:                types.Int64Type,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"concurrent_imports": {
						MarkdownDescription: "",
						Type:                types.Int64Type,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
				}),

				Required: true,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Marketo Connection identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (r *MarketoConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_marketo_connection"
}

type MarketoConnectionResource struct {
	client *polytomic.Client
}

func (r *MarketoConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:           data.Name.ValueString(),
			Type:           polytomic.MarketoConnectionType,
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.MarketoConfiguration{
				ClientID:          data.Configuration.Attributes()["client_id"].(types.String).ValueString(),
				ClientSecret:      data.Configuration.Attributes()["client_secret"].(types.String).ValueString(),
				RESTEndpoint:      data.Configuration.Attributes()["rest_endpoint"].(types.String).ValueString(),
				EnforceAPILimits:  data.Configuration.Attributes()["enforce_api_limits"].(types.Bool).ValueBool(),
				DailyAPICalls:     int(data.Configuration.Attributes()["daily_api_calls"].(types.Int64).ValueInt64()),
				ConcurrentImports: int(data.Configuration.Attributes()["concurrent_imports"].(types.Int64).ValueInt64()),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}
	data.Id = types.StringValue(created.ID)
	data.Name = types.StringValue(created.Name)

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Marketo", "id": created.ID})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *MarketoConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

func (r *MarketoConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
			Configuration: polytomic.MarketoConfiguration{
				ClientID:          data.Configuration.Attributes()["client_id"].(types.String).ValueString(),
				ClientSecret:      data.Configuration.Attributes()["client_secret"].(types.String).ValueString(),
				RESTEndpoint:      data.Configuration.Attributes()["rest_endpoint"].(types.String).ValueString(),
				EnforceAPILimits:  data.Configuration.Attributes()["enforce_api_limits"].(types.Bool).ValueBool(),
				DailyAPICalls:     int(data.Configuration.Attributes()["daily_api_calls"].(types.Int64).ValueInt64()),
				ConcurrentImports: int(data.Configuration.Attributes()["concurrent_imports"].(types.Int64).ValueInt64()),
			},
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

func (r *MarketoConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *MarketoConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MarketoConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
