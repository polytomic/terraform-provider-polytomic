// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package provider

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
	ptcore "github.com/polytomic/polytomic-go/core"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &DatabricksConnectionResource{}
var _ resource.ResourceWithImportState = &DatabricksConnectionResource{}

func (t *DatabricksConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Databricks Connection",
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
					"access_token": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           true,
					},
					"cloud_provider": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
					"enable_delta_uniform": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
					"enforce_query_limit": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
					"http_path": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"server_hostname": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"unity_catalog_enabled": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
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
				MarkdownDescription: "Databricks Connection identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

type DatabricksConf struct {
	Access_token string `mapstructure:"access_token" tfsdk:"access_token"`

	Cloud_provider string `mapstructure:"cloud_provider" tfsdk:"cloud_provider"`

	Enable_delta_uniform bool `mapstructure:"enable_delta_uniform" tfsdk:"enable_delta_uniform"`

	Enforce_query_limit bool `mapstructure:"enforce_query_limit" tfsdk:"enforce_query_limit"`

	Http_path string `mapstructure:"http_path" tfsdk:"http_path"`

	Port int64 `mapstructure:"port" tfsdk:"port"`

	Server_hostname string `mapstructure:"server_hostname" tfsdk:"server_hostname"`

	Unity_catalog_enabled bool `mapstructure:"unity_catalog_enabled" tfsdk:"unity_catalog_enabled"`
}

type DatabricksConnectionResource struct {
	provider *client.Provider
}

func (r *DatabricksConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *DatabricksConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databricks_connection"
}

func (r *DatabricksConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	created, err := client.Connections.Create(ctx, &polytomic.CreateConnectionRequestSchema{
		Name:           data.Name.ValueString(),
		Type:           "databricks",
		OrganizationId: data.Organization.ValueStringPointer(),
		Configuration: map[string]interface{}{
			"access_token":          data.Configuration.Attributes()["access_token"].(types.String).ValueString(),
			"cloud_provider":        data.Configuration.Attributes()["cloud_provider"].(types.String).ValueString(),
			"enable_delta_uniform":  data.Configuration.Attributes()["enable_delta_uniform"].(types.Bool).ValueBool(),
			"enforce_query_limit":   data.Configuration.Attributes()["enforce_query_limit"].(types.Bool).ValueBool(),
			"http_path":             data.Configuration.Attributes()["http_path"].(types.String).ValueString(),
			"port":                  int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
			"server_hostname":       data.Configuration.Attributes()["server_hostname"].(types.String).ValueString(),
			"unity_catalog_enabled": data.Configuration.Attributes()["unity_catalog_enabled"].(types.Bool).ValueBool(),
		},
		Validate: pointer.ToBool(false),
	})
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}
	data.Id = types.StringPointerValue(created.Data.Id)
	data.Name = types.StringPointerValue(created.Data.Name)
	data.Organization = types.StringPointerValue(created.Data.OrganizationId)

	conf := DatabricksConf{}
	err = mapstructure.Decode(created.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"access_token":          types.StringType,
		"cloud_provider":        types.StringType,
		"enable_delta_uniform":  types.BoolType,
		"enforce_query_limit":   types.BoolType,
		"http_path":             types.StringType,
		"port":                  types.NumberType,
		"server_hostname":       types.StringType,
		"unity_catalog_enabled": types.BoolType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Databricks", "id": created.Data.Id})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *DatabricksConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading connection: %s", err))
		return
	}
	data.Id = types.StringPointerValue(connection.Data.Id)
	data.Name = types.StringPointerValue(connection.Data.Name)
	data.Organization = types.StringPointerValue(connection.Data.OrganizationId)

	conf := DatabricksConf{}
	err = mapstructure.Decode(connection.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"access_token":          types.StringType,
		"cloud_provider":        types.StringType,
		"enable_delta_uniform":  types.BoolType,
		"enforce_query_limit":   types.BoolType,
		"http_path":             types.StringType,
		"port":                  types.NumberType,
		"server_hostname":       types.StringType,
		"unity_catalog_enabled": types.BoolType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *DatabricksConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	updated, err := client.Connections.Update(ctx,
		data.Id.ValueString(),
		&polytomic.UpdateConnectionRequestSchema{
			Name:           data.Name.ValueString(),
			OrganizationId: data.Organization.ValueStringPointer(),
			Configuration: map[string]interface{}{
				"access_token":          data.Configuration.Attributes()["access_token"].(types.String).ValueString(),
				"cloud_provider":        data.Configuration.Attributes()["cloud_provider"].(types.String).ValueString(),
				"enable_delta_uniform":  data.Configuration.Attributes()["enable_delta_uniform"].(types.Bool).ValueBool(),
				"enforce_query_limit":   data.Configuration.Attributes()["enforce_query_limit"].(types.Bool).ValueBool(),
				"http_path":             data.Configuration.Attributes()["http_path"].(types.String).ValueString(),
				"port":                  int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
				"server_hostname":       data.Configuration.Attributes()["server_hostname"].(types.String).ValueString(),
				"unity_catalog_enabled": data.Configuration.Attributes()["unity_catalog_enabled"].(types.Bool).ValueBool(),
			},
			Validate: pointer.ToBool(false),
		})
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	data.Id = types.StringPointerValue(updated.Data.Id)
	data.Name = types.StringPointerValue(updated.Data.Name)
	data.Organization = types.StringPointerValue(updated.Data.OrganizationId)

	conf := DatabricksConf{}
	err = mapstructure.Decode(updated.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"access_token":          types.StringType,
		"cloud_provider":        types.StringType,
		"enable_delta_uniform":  types.BoolType,
		"enforce_query_limit":   types.BoolType,
		"http_path":             types.StringType,
		"port":                  types.NumberType,
		"server_hostname":       types.StringType,
		"unity_catalog_enabled": types.BoolType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *DatabricksConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

			resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting connection: %s", err))
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
			resp.Diagnostics.AddError("Connection in use",
				fmt.Sprintf("Connection is used by %s \"%s\" (%s). Please remove before deleting this connection.",
					pErr.Body.Metadata["type"], pErr.Body.Metadata["name"], pErr.Body.Metadata["id"]),
			)
			return
		}
	}

	resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting connection: %s", err))

}

func (r *DatabricksConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
