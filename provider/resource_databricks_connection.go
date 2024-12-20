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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
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
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"auth_mode": schema.StringAttribute{
						MarkdownDescription: "How to authenticate with AWS. Defaults to Access Key and Secret",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"aws_access_key_id": schema.StringAttribute{
						MarkdownDescription: "See https://docs.polytomic.com/docs/databricks-connections#writing-to-databricks",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"aws_secret_access_key": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"aws_user": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"azure_access_key": schema.StringAttribute{
						MarkdownDescription: "The access key associated with this storage account",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"azure_account_name": schema.StringAttribute{
						MarkdownDescription: "The account name of the storage account",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"cloud_provider": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"concurrent_queries": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"container_name": schema.StringAttribute{
						MarkdownDescription: "The container which we will stage files in",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"deleted_file_retention_days": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"enable_delta_uniform": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"enforce_query_limit": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"external_id": schema.StringAttribute{
						MarkdownDescription: "External ID for the IAM role",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"http_path": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"iam_role_arn": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"log_file_retention_days": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"s3_bucket_name": schema.StringAttribute{
						MarkdownDescription: "Name of bucket used for staging data load files",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"s3_bucket_region": schema.StringAttribute{
						MarkdownDescription: "Region of bucket.example=us-east-1",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"server_hostname": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"set_retention_properties": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"storage_credential_name": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"unity_catalog_enabled": schema.BoolAttribute{
						MarkdownDescription: "",
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

	Auth_mode string `mapstructure:"auth_mode" tfsdk:"auth_mode"`

	Aws_access_key_id string `mapstructure:"aws_access_key_id" tfsdk:"aws_access_key_id"`

	Aws_secret_access_key string `mapstructure:"aws_secret_access_key" tfsdk:"aws_secret_access_key"`

	Aws_user string `mapstructure:"aws_user" tfsdk:"aws_user"`

	Azure_access_key string `mapstructure:"azure_access_key" tfsdk:"azure_access_key"`

	Azure_account_name string `mapstructure:"azure_account_name" tfsdk:"azure_account_name"`

	Cloud_provider string `mapstructure:"cloud_provider" tfsdk:"cloud_provider"`

	Concurrent_queries int64 `mapstructure:"concurrent_queries" tfsdk:"concurrent_queries"`

	Container_name string `mapstructure:"container_name" tfsdk:"container_name"`

	Deleted_file_retention_days int64 `mapstructure:"deleted_file_retention_days" tfsdk:"deleted_file_retention_days"`

	Enable_delta_uniform bool `mapstructure:"enable_delta_uniform" tfsdk:"enable_delta_uniform"`

	Enforce_query_limit bool `mapstructure:"enforce_query_limit" tfsdk:"enforce_query_limit"`

	External_id string `mapstructure:"external_id" tfsdk:"external_id"`

	Http_path string `mapstructure:"http_path" tfsdk:"http_path"`

	Iam_role_arn string `mapstructure:"iam_role_arn" tfsdk:"iam_role_arn"`

	Log_file_retention_days int64 `mapstructure:"log_file_retention_days" tfsdk:"log_file_retention_days"`

	Port int64 `mapstructure:"port" tfsdk:"port"`

	S3_bucket_name string `mapstructure:"s3_bucket_name" tfsdk:"s3_bucket_name"`

	S3_bucket_region string `mapstructure:"s3_bucket_region" tfsdk:"s3_bucket_region"`

	Server_hostname string `mapstructure:"server_hostname" tfsdk:"server_hostname"`

	Set_retention_properties bool `mapstructure:"set_retention_properties" tfsdk:"set_retention_properties"`

	Storage_credential_name string `mapstructure:"storage_credential_name" tfsdk:"storage_credential_name"`

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
			"access_token":                data.Configuration.Attributes()["access_token"].(types.String).ValueString(),
			"auth_mode":                   data.Configuration.Attributes()["auth_mode"].(types.String).ValueString(),
			"aws_access_key_id":           data.Configuration.Attributes()["aws_access_key_id"].(types.String).ValueString(),
			"aws_secret_access_key":       data.Configuration.Attributes()["aws_secret_access_key"].(types.String).ValueString(),
			"aws_user":                    data.Configuration.Attributes()["aws_user"].(types.String).ValueString(),
			"azure_access_key":            data.Configuration.Attributes()["azure_access_key"].(types.String).ValueString(),
			"azure_account_name":          data.Configuration.Attributes()["azure_account_name"].(types.String).ValueString(),
			"cloud_provider":              data.Configuration.Attributes()["cloud_provider"].(types.String).ValueString(),
			"concurrent_queries":          int(data.Configuration.Attributes()["concurrent_queries"].(types.Int64).ValueInt64()),
			"container_name":              data.Configuration.Attributes()["container_name"].(types.String).ValueString(),
			"deleted_file_retention_days": int(data.Configuration.Attributes()["deleted_file_retention_days"].(types.Int64).ValueInt64()),
			"enable_delta_uniform":        data.Configuration.Attributes()["enable_delta_uniform"].(types.Bool).ValueBool(),
			"enforce_query_limit":         data.Configuration.Attributes()["enforce_query_limit"].(types.Bool).ValueBool(),
			"external_id":                 data.Configuration.Attributes()["external_id"].(types.String).ValueString(),
			"http_path":                   data.Configuration.Attributes()["http_path"].(types.String).ValueString(),
			"iam_role_arn":                data.Configuration.Attributes()["iam_role_arn"].(types.String).ValueString(),
			"log_file_retention_days":     int(data.Configuration.Attributes()["log_file_retention_days"].(types.Int64).ValueInt64()),
			"port":                        int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
			"s3_bucket_name":              data.Configuration.Attributes()["s3_bucket_name"].(types.String).ValueString(),
			"s3_bucket_region":            data.Configuration.Attributes()["s3_bucket_region"].(types.String).ValueString(),
			"server_hostname":             data.Configuration.Attributes()["server_hostname"].(types.String).ValueString(),
			"set_retention_properties":    data.Configuration.Attributes()["set_retention_properties"].(types.Bool).ValueBool(),
			"storage_credential_name":     data.Configuration.Attributes()["storage_credential_name"].(types.String).ValueString(),
			"unity_catalog_enabled":       data.Configuration.Attributes()["unity_catalog_enabled"].(types.Bool).ValueBool(),
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
		"access_token":                types.StringType,
		"auth_mode":                   types.StringType,
		"aws_access_key_id":           types.StringType,
		"aws_secret_access_key":       types.StringType,
		"aws_user":                    types.StringType,
		"azure_access_key":            types.StringType,
		"azure_account_name":          types.StringType,
		"cloud_provider":              types.StringType,
		"concurrent_queries":          types.NumberType,
		"container_name":              types.StringType,
		"deleted_file_retention_days": types.NumberType,
		"enable_delta_uniform":        types.BoolType,
		"enforce_query_limit":         types.BoolType,
		"external_id":                 types.StringType,
		"http_path":                   types.StringType,
		"iam_role_arn":                types.StringType,
		"log_file_retention_days":     types.NumberType,
		"port":                        types.NumberType,
		"s3_bucket_name":              types.StringType,
		"s3_bucket_region":            types.StringType,
		"server_hostname":             types.StringType,
		"set_retention_properties":    types.BoolType,
		"storage_credential_name":     types.StringType,
		"unity_catalog_enabled":       types.BoolType,
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
		"access_token":                types.StringType,
		"auth_mode":                   types.StringType,
		"aws_access_key_id":           types.StringType,
		"aws_secret_access_key":       types.StringType,
		"aws_user":                    types.StringType,
		"azure_access_key":            types.StringType,
		"azure_account_name":          types.StringType,
		"cloud_provider":              types.StringType,
		"concurrent_queries":          types.NumberType,
		"container_name":              types.StringType,
		"deleted_file_retention_days": types.NumberType,
		"enable_delta_uniform":        types.BoolType,
		"enforce_query_limit":         types.BoolType,
		"external_id":                 types.StringType,
		"http_path":                   types.StringType,
		"iam_role_arn":                types.StringType,
		"log_file_retention_days":     types.NumberType,
		"port":                        types.NumberType,
		"s3_bucket_name":              types.StringType,
		"s3_bucket_region":            types.StringType,
		"server_hostname":             types.StringType,
		"set_retention_properties":    types.BoolType,
		"storage_credential_name":     types.StringType,
		"unity_catalog_enabled":       types.BoolType,
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
				"access_token":                data.Configuration.Attributes()["access_token"].(types.String).ValueString(),
				"auth_mode":                   data.Configuration.Attributes()["auth_mode"].(types.String).ValueString(),
				"aws_access_key_id":           data.Configuration.Attributes()["aws_access_key_id"].(types.String).ValueString(),
				"aws_secret_access_key":       data.Configuration.Attributes()["aws_secret_access_key"].(types.String).ValueString(),
				"aws_user":                    data.Configuration.Attributes()["aws_user"].(types.String).ValueString(),
				"azure_access_key":            data.Configuration.Attributes()["azure_access_key"].(types.String).ValueString(),
				"azure_account_name":          data.Configuration.Attributes()["azure_account_name"].(types.String).ValueString(),
				"cloud_provider":              data.Configuration.Attributes()["cloud_provider"].(types.String).ValueString(),
				"concurrent_queries":          int(data.Configuration.Attributes()["concurrent_queries"].(types.Int64).ValueInt64()),
				"container_name":              data.Configuration.Attributes()["container_name"].(types.String).ValueString(),
				"deleted_file_retention_days": int(data.Configuration.Attributes()["deleted_file_retention_days"].(types.Int64).ValueInt64()),
				"enable_delta_uniform":        data.Configuration.Attributes()["enable_delta_uniform"].(types.Bool).ValueBool(),
				"enforce_query_limit":         data.Configuration.Attributes()["enforce_query_limit"].(types.Bool).ValueBool(),
				"external_id":                 data.Configuration.Attributes()["external_id"].(types.String).ValueString(),
				"http_path":                   data.Configuration.Attributes()["http_path"].(types.String).ValueString(),
				"iam_role_arn":                data.Configuration.Attributes()["iam_role_arn"].(types.String).ValueString(),
				"log_file_retention_days":     int(data.Configuration.Attributes()["log_file_retention_days"].(types.Int64).ValueInt64()),
				"port":                        int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
				"s3_bucket_name":              data.Configuration.Attributes()["s3_bucket_name"].(types.String).ValueString(),
				"s3_bucket_region":            data.Configuration.Attributes()["s3_bucket_region"].(types.String).ValueString(),
				"server_hostname":             data.Configuration.Attributes()["server_hostname"].(types.String).ValueString(),
				"set_retention_properties":    data.Configuration.Attributes()["set_retention_properties"].(types.Bool).ValueBool(),
				"storage_credential_name":     data.Configuration.Attributes()["storage_credential_name"].(types.String).ValueString(),
				"unity_catalog_enabled":       data.Configuration.Attributes()["unity_catalog_enabled"].(types.Bool).ValueBool(),
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
		"access_token":                types.StringType,
		"auth_mode":                   types.StringType,
		"aws_access_key_id":           types.StringType,
		"aws_secret_access_key":       types.StringType,
		"aws_user":                    types.StringType,
		"azure_access_key":            types.StringType,
		"azure_account_name":          types.StringType,
		"cloud_provider":              types.StringType,
		"concurrent_queries":          types.NumberType,
		"container_name":              types.StringType,
		"deleted_file_retention_days": types.NumberType,
		"enable_delta_uniform":        types.BoolType,
		"enforce_query_limit":         types.BoolType,
		"external_id":                 types.StringType,
		"http_path":                   types.StringType,
		"iam_role_arn":                types.StringType,
		"log_file_retention_days":     types.NumberType,
		"port":                        types.NumberType,
		"s3_bucket_name":              types.StringType,
		"s3_bucket_region":            types.StringType,
		"server_hostname":             types.StringType,
		"set_retention_properties":    types.BoolType,
		"storage_credential_name":     types.StringType,
		"unity_catalog_enabled":       types.BoolType,
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
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting connection: %s", err))
	}
}

func (r *DatabricksConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
