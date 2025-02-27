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
var _ resource.Resource = &PardotConnectionResource{}
var _ resource.ResourceWithImportState = &PardotConnectionResource{}

var PardotSchema = schema.Schema{
	MarkdownDescription: ":meta:subcategory:Connections: Pardot Connection",
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
				"account_type": schema.StringAttribute{
					MarkdownDescription: `Account type`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"business_unit_id": schema.StringAttribute{
					MarkdownDescription: `Business Unit ID`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"daily_api_calls": schema.Int64Attribute{
					MarkdownDescription: `Daily call limit`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"enforce_api_limits": schema.BoolAttribute{
					MarkdownDescription: `Enforce API limits`,
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
			MarkdownDescription: "Pardot Connection identifier",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	},
}

func (t *PardotConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PardotSchema
}

type PardotConf struct {
	Account_type       string `mapstructure:"account_type" tfsdk:"account_type"`
	Business_unit_id   string `mapstructure:"business_unit_id" tfsdk:"business_unit_id"`
	Daily_api_calls    int64  `mapstructure:"daily_api_calls" tfsdk:"daily_api_calls"`
	Enforce_api_limits bool   `mapstructure:"enforce_api_limits" tfsdk:"enforce_api_limits"`
	Username           string `mapstructure:"username" tfsdk:"username"`
}

type PardotConnectionResource struct {
	provider *providerclient.Provider
}

func (r *PardotConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *PardotConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pardot_connection"
}

func (r *PardotConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
		Type:           "pardot",
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

	conf := PardotConf{}
	err = mapstructure.Decode(created.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"account_type":       types.StringType,
		"business_unit_id":   types.StringType,
		"daily_api_calls":    types.NumberType,
		"enforce_api_limits": types.BoolType,
		"username":           types.StringType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Pardot", "id": created.Data.Id})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *PardotConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	configAttributes, ok := getConfigAttributes(PardotSchema)
	if !ok {
		resp.Diagnostics.AddError("Error getting connection configuration attributes", "Could not get configuration attributes")
		return
	}

	originalConfData, err := objectMapValue(ctx, data.Configuration)
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection configuration", err.Error())
		return
	}

	// reset sensitive values so terraform doesn't think we have changes
	connection.Data.Configuration = resetSensitiveValues(configAttributes, originalConfData, connection.Data.Configuration)

	conf := PardotConf{}
	err = mapstructure.Decode(connection.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"account_type":       types.StringType,
		"business_unit_id":   types.StringType,
		"daily_api_calls":    types.NumberType,
		"enforce_api_limits": types.BoolType,
		"username":           types.StringType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *PardotConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	configAttributes, ok := getConfigAttributes(PardotSchema)
	if !ok {
		resp.Diagnostics.AddError("Error getting connection configuration attributes", "Could not get configuration attributes")
		return
	}

	var prevData connectionData

	diags = req.State.Get(ctx, &prevData)
	resp.Diagnostics.Append(diags...)

	connConf = handleSensitiveValues(ctx, configAttributes, connConf, prevData.Configuration.Attributes())

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

	conf := PardotConf{}
	err = mapstructure.Decode(updated.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"account_type":       types.StringType,
		"business_unit_id":   types.StringType,
		"daily_api_calls":    types.NumberType,
		"enforce_api_limits": types.BoolType,
		"username":           types.StringType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *PardotConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *PardotConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
