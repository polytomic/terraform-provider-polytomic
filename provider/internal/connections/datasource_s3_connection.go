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
var _ datasource.DataSource = &S3ConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type S3ConnectionDataSource struct {
	provider *providerclient.Provider
}

func (d *S3ConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *S3ConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_connection"
}

func (d *S3ConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: S3 Connection",
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
						MarkdownDescription: `AWS Access Key ID

    Access Key ID with read/write access to a bucket.`,
						Computed: true,
					},
					"aws_user": schema.StringAttribute{
						MarkdownDescription: `User ARN`,
						Computed:            true,
					},
					"external_id": schema.StringAttribute{
						MarkdownDescription: `External ID for the IAM role`,
						Computed:            true,
					},
					"iam_role_arn": schema.StringAttribute{
						MarkdownDescription: `IAM Role ARN`,
						Computed:            true,
					},
					"is_single_table": schema.BoolAttribute{
						MarkdownDescription: `Files are time-based snapshots

    Treat the files as a single table.`,
						Computed: true,
					},
					"s3_bucket_name": schema.StringAttribute{
						MarkdownDescription: `S3 Bucket Name

    Bucket name (folder optional); ex: s3://polytomic/dataset`,
						Computed: true,
					},
					"s3_bucket_region": schema.StringAttribute{
						MarkdownDescription: `S3 Bucket Region`,
						Computed:            true,
					},
					"single_table_file_format": schema.StringAttribute{
						MarkdownDescription: `File format`,
						Computed:            true,
					},
					"single_table_name": schema.StringAttribute{
						MarkdownDescription: `Collection name`,
						Computed:            true,
					},
					"skip_lines": schema.Int64Attribute{
						MarkdownDescription: `Skip first lines

    Skip first N lines of each CSV file.`,
						Computed: true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (d *S3ConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
			"external_id": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["external_id"], "string").(string),
			),
			"iam_role_arn": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["iam_role_arn"], "string").(string),
			),
			"is_single_table": types.BoolValue(
				getValueOrEmpty(connection.Data.Configuration["is_single_table"], "bool").(bool),
			),
			"s3_bucket_name": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["s3_bucket_name"], "string").(string),
			),
			"s3_bucket_region": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["s3_bucket_region"], "string").(string),
			),
			"single_table_file_format": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["single_table_file_format"], "string").(string),
			),
			"single_table_name": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["single_table_name"], "string").(string),
			),
			"skip_lines": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["skip_lines"], "string").(string),
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
