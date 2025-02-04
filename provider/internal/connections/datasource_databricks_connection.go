// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package connections

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/providerclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &DatabricksConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type DatabricksConnectionDataSource struct {
	provider *providerclient.Provider
}

func (d *DatabricksConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *DatabricksConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databricks_connection"
}

func (d *DatabricksConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Databricks Connection",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"configuration": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"auth_mode": schema.StringAttribute{
						MarkdownDescription: `Authentication Method

    How to authenticate with AWS. Defaults to Access Key and Secret`,
						Computed: true,
					},
					"aws_access_key_id": schema.StringAttribute{
						MarkdownDescription: `AWS Access Key ID (destinations only)

    See https://docs.polytomic.com/docs/databricks-connections#writing-to-databricks`,
						Computed: true,
					},
					"aws_user": schema.StringAttribute{
						MarkdownDescription: `User ARN (destinations only)`,
						Computed:            true,
					},
					"azure_account_name": schema.StringAttribute{
						MarkdownDescription: `Storage Account Name (destination support only)

    The account name of the storage account`,
						Computed: true,
					},
					"cloud_provider": schema.StringAttribute{
						MarkdownDescription: `Cloud Provider (destination support only)`,
						Computed:            true,
					},
					"concurrent_queries": schema.Int64Attribute{
						MarkdownDescription: `Concurrent query limit`,
						Computed:            true,
					},
					"container_name": schema.StringAttribute{
						MarkdownDescription: `Storage Container Name (destination support only)

    The container which we will stage files in`,
						Computed: true,
					},
					"deleted_file_retention_days": schema.Int64Attribute{
						MarkdownDescription: `Deleted file retention`,
						Computed:            true,
					},
					"enable_delta_uniform": schema.BoolAttribute{
						MarkdownDescription: `Enable Delta UniForm tables`,
						Computed:            true,
					},
					"enforce_query_limit": schema.BoolAttribute{
						MarkdownDescription: `Limit concurrent queries`,
						Computed:            true,
					},
					"external_id": schema.StringAttribute{
						MarkdownDescription: `External ID

    External ID for the IAM role`,
						Computed: true,
					},
					"http_path": schema.StringAttribute{
						MarkdownDescription: `HTTP Path`,
						Computed:            true,
					},
					"iam_role_arn": schema.StringAttribute{
						MarkdownDescription: `IAM Role ARN`,
						Computed:            true,
					},
					"log_file_retention_days": schema.Int64Attribute{
						MarkdownDescription: `Log retention`,
						Computed:            true,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: ``,
						Computed:            true,
					},
					"s3_bucket_name": schema.StringAttribute{
						MarkdownDescription: `S3 Bucket Name (destinations only)

    Name of bucket used for staging data load files`,
						Computed: true,
					},
					"s3_bucket_region": schema.StringAttribute{
						MarkdownDescription: `S3 Bucket Region (destinations only)

    Region of bucket`,
						Computed: true,
					},
					"server_hostname": schema.StringAttribute{
						MarkdownDescription: `Server Hostname`,
						Computed:            true,
					},
					"set_retention_properties": schema.BoolAttribute{
						MarkdownDescription: `Configure data retention for tables`,
						Computed:            true,
					},
					"storage_credential_name": schema.StringAttribute{
						MarkdownDescription: `Storage credential name`,
						Computed:            true,
					},
					"unity_catalog_enabled": schema.BoolAttribute{
						MarkdownDescription: `Unity Catalog enabled`,
						Computed:            true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (d *DatabricksConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data connectionData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the connection
	client, err := d.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	connection, err := client.Connections.Get(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection", err.Error())
		return
	}

	data.Id = types.StringPointerValue(connection.Data.Id)
	data.Name = types.StringPointerValue(connection.Data.Name)
	data.Organization = types.StringPointerValue(connection.Data.OrganizationId)
	var diags diag.Diagnostics
	data.Configuration, diags = types.ObjectValue(
		data.Configuration.AttributeTypes(ctx),
		map[string]attr.Value{
			"auth_mode": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["auth_mode"], "string").(string),
			),
			"aws_access_key_id": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["aws_access_key_id"], "string").(string),
			),
			"aws_user": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["aws_user"], "string").(string),
			),
			"azure_account_name": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["azure_account_name"], "string").(string),
			),
			"cloud_provider": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["cloud_provider"], "string").(string),
			),
			"concurrent_queries": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["concurrent_queries"], "string").(string),
			),
			"container_name": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["container_name"], "string").(string),
			),
			"deleted_file_retention_days": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["deleted_file_retention_days"], "string").(string),
			),
			"enable_delta_uniform": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["enable_delta_uniform"], "bool").(bool),
			),
			"enforce_query_limit": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["enforce_query_limit"], "bool").(bool),
			),
			"external_id": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["external_id"], "string").(string),
			),
			"http_path": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["http_path"], "string").(string),
			),
			"iam_role_arn": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["iam_role_arn"], "string").(string),
			),
			"log_file_retention_days": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["log_file_retention_days"], "string").(string),
			),
			"port": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["port"], "string").(string),
			),
			"s3_bucket_name": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["s3_bucket_name"], "string").(string),
			),
			"s3_bucket_region": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["s3_bucket_region"], "string").(string),
			),
			"server_hostname": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["server_hostname"], "string").(string),
			),
			"set_retention_properties": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["set_retention_properties"], "bool").(bool),
			),
			"storage_credential_name": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["storage_credential_name"], "string").(string),
			),
			"unity_catalog_enabled": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["unity_catalog_enabled"], "bool").(bool),
			),
		},
	)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
