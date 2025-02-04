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
var _ resource.Resource = &DatabricksConnectionResource{}
var _ resource.ResourceWithImportState = &DatabricksConnectionResource{}

var DatabricksSchema = schema.Schema{
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
					MarkdownDescription: `Access Token`,
					Required:            true,
					Optional:            false,
					Computed:            false,
					Sensitive:           true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"auth_mode": schema.StringAttribute{
					MarkdownDescription: `Authentication Method

    How to authenticate with AWS. Defaults to Access Key and Secret`,
					Required:  true,
					Optional:  false,
					Computed:  false,
					Sensitive: false,
				},
				"aws_access_key_id": schema.StringAttribute{
					MarkdownDescription: `AWS Access Key ID (destinations only)

    See https://docs.polytomic.com/docs/databricks-connections#writing-to-databricks`,
					Required:  false,
					Optional:  true,
					Computed:  true,
					Sensitive: false,
				},
				"aws_secret_access_key": schema.StringAttribute{
					MarkdownDescription: `AWS Secret Access Key (destinations only)`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"aws_user": schema.StringAttribute{
					MarkdownDescription: `User ARN (destinations only)`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"azure_access_key": schema.StringAttribute{
					MarkdownDescription: `Storage Account Access Key (destination support only)

    The access key associated with this storage account`,
					Required:  false,
					Optional:  true,
					Computed:  true,
					Sensitive: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"azure_account_name": schema.StringAttribute{
					MarkdownDescription: `Storage Account Name (destination support only)

    The account name of the storage account`,
					Required:  false,
					Optional:  true,
					Computed:  true,
					Sensitive: false,
				},
				"cloud_provider": schema.StringAttribute{
					MarkdownDescription: `Cloud Provider (destination support only)`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"concurrent_queries": schema.Int64Attribute{
					MarkdownDescription: `Concurrent query limit`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"container_name": schema.StringAttribute{
					MarkdownDescription: `Storage Container Name (destination support only)

    The container which we will stage files in`,
					Required:  false,
					Optional:  true,
					Computed:  true,
					Sensitive: false,
				},
				"deleted_file_retention_days": schema.Int64Attribute{
					MarkdownDescription: `Deleted file retention`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"enable_delta_uniform": schema.BoolAttribute{
					MarkdownDescription: `Enable Delta UniForm tables`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"enforce_query_limit": schema.BoolAttribute{
					MarkdownDescription: `Limit concurrent queries`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"external_id": schema.StringAttribute{
					MarkdownDescription: `External ID

    External ID for the IAM role`,
					Required:  false,
					Optional:  true,
					Computed:  true,
					Sensitive: false,
				},
				"http_path": schema.StringAttribute{
					MarkdownDescription: `HTTP Path`,
					Required:            true,
					Optional:            false,
					Computed:            false,
					Sensitive:           false,
				},
				"iam_role_arn": schema.StringAttribute{
					MarkdownDescription: `IAM Role ARN`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"log_file_retention_days": schema.Int64Attribute{
					MarkdownDescription: `Log retention`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"port": schema.Int64Attribute{
					MarkdownDescription: ``,
					Required:            true,
					Optional:            false,
					Computed:            false,
					Sensitive:           false,
				},
				"s3_bucket_name": schema.StringAttribute{
					MarkdownDescription: `S3 Bucket Name (destinations only)

    Name of bucket used for staging data load files`,
					Required:  false,
					Optional:  true,
					Computed:  true,
					Sensitive: false,
				},
				"s3_bucket_region": schema.StringAttribute{
					MarkdownDescription: `S3 Bucket Region (destinations only)

    Region of bucket.example=us-east-1`,
					Required:  false,
					Optional:  true,
					Computed:  true,
					Sensitive: false,
				},
				"server_hostname": schema.StringAttribute{
					MarkdownDescription: `Server Hostname`,
					Required:            true,
					Optional:            false,
					Computed:            false,
					Sensitive:           false,
				},
				"set_retention_properties": schema.BoolAttribute{
					MarkdownDescription: `Configure data retention for tables`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"storage_credential_name": schema.StringAttribute{
					MarkdownDescription: `Storage credential name`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"unity_catalog_enabled": schema.BoolAttribute{
					MarkdownDescription: `Unity Catalog enabled`,
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

func (t *DatabricksConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DatabricksSchema
}

type DatabricksConf struct {
	Access_token                string `mapstructure:"access_token" tfsdk:"access_token"`
	Auth_mode                   string `mapstructure:"auth_mode" tfsdk:"auth_mode"`
	Aws_access_key_id           string `mapstructure:"aws_access_key_id" tfsdk:"aws_access_key_id"`
	Aws_secret_access_key       string `mapstructure:"aws_secret_access_key" tfsdk:"aws_secret_access_key"`
	Aws_user                    string `mapstructure:"aws_user" tfsdk:"aws_user"`
	Azure_access_key            string `mapstructure:"azure_access_key" tfsdk:"azure_access_key"`
	Azure_account_name          string `mapstructure:"azure_account_name" tfsdk:"azure_account_name"`
	Cloud_provider              string `mapstructure:"cloud_provider" tfsdk:"cloud_provider"`
	Concurrent_queries          int64  `mapstructure:"concurrent_queries" tfsdk:"concurrent_queries"`
	Container_name              string `mapstructure:"container_name" tfsdk:"container_name"`
	Deleted_file_retention_days int64  `mapstructure:"deleted_file_retention_days" tfsdk:"deleted_file_retention_days"`
	Enable_delta_uniform        bool   `mapstructure:"enable_delta_uniform" tfsdk:"enable_delta_uniform"`
	Enforce_query_limit         bool   `mapstructure:"enforce_query_limit" tfsdk:"enforce_query_limit"`
	External_id                 string `mapstructure:"external_id" tfsdk:"external_id"`
	Http_path                   string `mapstructure:"http_path" tfsdk:"http_path"`
	Iam_role_arn                string `mapstructure:"iam_role_arn" tfsdk:"iam_role_arn"`
	Log_file_retention_days     int64  `mapstructure:"log_file_retention_days" tfsdk:"log_file_retention_days"`
	Port                        int64  `mapstructure:"port" tfsdk:"port"`
	S3_bucket_name              string `mapstructure:"s3_bucket_name" tfsdk:"s3_bucket_name"`
	S3_bucket_region            string `mapstructure:"s3_bucket_region" tfsdk:"s3_bucket_region"`
	Server_hostname             string `mapstructure:"server_hostname" tfsdk:"server_hostname"`
	Set_retention_properties    bool   `mapstructure:"set_retention_properties" tfsdk:"set_retention_properties"`
	Storage_credential_name     string `mapstructure:"storage_credential_name" tfsdk:"storage_credential_name"`
	Unity_catalog_enabled       bool   `mapstructure:"unity_catalog_enabled" tfsdk:"unity_catalog_enabled"`
}

type DatabricksConnectionResource struct {
	provider *providerclient.Provider
}

func (r *DatabricksConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
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
	connConf, err := objectMapValue(ctx, data.Configuration)
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection configuration", err.Error())
		return
	}
	created, err := client.Connections.Create(ctx, &polytomic.CreateConnectionRequestSchema{
		Name:           data.Name.ValueString(),
		Type:           "databricks",
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

	conf := DatabricksConf{}
	err = mapstructure.Decode(created.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
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
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error reading connection: %s", err))
		return
	}
	data.Id = types.StringPointerValue(connection.Data.Id)
	data.Name = types.StringPointerValue(connection.Data.Name)
	data.Organization = types.StringPointerValue(connection.Data.OrganizationId)

	configAttributes, ok := getConfigAttributes(DatabricksSchema)
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

	conf := DatabricksConf{}
	err = mapstructure.Decode(connection.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
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
	connConf, err := objectMapValue(ctx, data.Configuration)
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection configuration", err.Error())
		return
	}

	configAttributes, ok := getConfigAttributes(DatabricksSchema)
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

	conf := DatabricksConf{}
	err = mapstructure.Decode(updated.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
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

func (r *DatabricksConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
