// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
var _ resource.Resource = &RedshiftConnectionResource{}
var _ resource.ResourceWithImportState = &RedshiftConnectionResource{}

func (t *RedshiftConnectionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Redshift Connection",
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
					"hostname": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"username": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"password": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           true,
					},
					"database": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"port": {
						MarkdownDescription: "",
						Type:                types.Int64Type,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh": {
						MarkdownDescription: "",
						Type:                types.BoolType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh_user": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh_host": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh_port": {
						MarkdownDescription: "",
						Type:                types.Int64Type,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh_private_key": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           true,
					},
					"aws_access_key_id": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"aws_secret_access_key": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           true,
					},
					"s3_bucket_name": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"s3_bucket_region": {
						MarkdownDescription: "",
						Type:                types.StringType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
				}),

				Required: true,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Redshift Connection identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (r *RedshiftConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redshift_connection"
}

type RedshiftConnectionResource struct {
	client *polytomic.Client
}

func (r *RedshiftConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:           data.Name.ValueString(),
			Type:           polytomic.RedshiftConnectionType,
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.RedshiftConnectionConfiguration{
				Hostname:           data.Configuration.Attributes()["hostname"].(types.String).ValueString(),
				Username:           data.Configuration.Attributes()["username"].(types.String).ValueString(),
				Password:           data.Configuration.Attributes()["password"].(types.String).ValueString(),
				Database:           data.Configuration.Attributes()["database"].(types.String).ValueString(),
				Port:               int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
				SSH:                data.Configuration.Attributes()["ssh"].(types.Bool).ValueBool(),
				SSHUser:            data.Configuration.Attributes()["ssh_user"].(types.String).ValueString(),
				SSHHost:            data.Configuration.Attributes()["ssh_host"].(types.String).ValueString(),
				SSHPort:            int(data.Configuration.Attributes()["ssh_port"].(types.Int64).ValueInt64()),
				SSHPrivateKey:      data.Configuration.Attributes()["ssh_private_key"].(types.String).ValueString(),
				AwsAccessKeyID:     data.Configuration.Attributes()["aws_access_key_id"].(types.String).ValueString(),
				AwsSecretAccessKey: data.Configuration.Attributes()["aws_secret_access_key"].(types.String).ValueString(),
				S3BucketName:       data.Configuration.Attributes()["s3_bucket_name"].(types.String).ValueString(),
				S3BucketRegion:     data.Configuration.Attributes()["s3_bucket_region"].(types.String).ValueString(),
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

	//var output polytomic.RedshiftConnectionConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(created.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"hostname": types.StringType,
	//
	//	"username": types.StringType,
	//
	//	"password": types.StringType,
	//
	//	"database": types.StringType,
	//
	//	"port": types.Int64Type,
	//
	//	"ssh": types.BoolType,
	//
	//	"ssh_user": types.StringType,
	//
	//	"ssh_host": types.StringType,
	//
	//	"ssh_port": types.Int64Type,
	//
	//	"ssh_private_key": types.StringType,
	//
	//	"aws_access_key_id": types.StringType,
	//
	//	"aws_secret_access_key": types.StringType,
	//
	//	"s3_bucket_name": types.StringType,
	//
	//	"s3_bucket_region": types.StringType,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Redshift", "id": created.ID})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *RedshiftConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
		}
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading connection: %s", err))
		return
	}

	data.Id = types.StringValue(connection.ID)
	data.Name = types.StringValue(connection.Name)
	data.Organization = types.StringValue(connection.OrganizationId)

	//var output polytomic.RedshiftConnectionConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(connection.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"hostname": types.StringType,
	//
	//	"username": types.StringType,
	//
	//	"password": types.StringType,
	//
	//	"database": types.StringType,
	//
	//	"port": types.Int64Type,
	//
	//	"ssh": types.BoolType,
	//
	//	"ssh_user": types.StringType,
	//
	//	"ssh_host": types.StringType,
	//
	//	"ssh_port": types.Int64Type,
	//
	//	"ssh_private_key": types.StringType,
	//
	//	"aws_access_key_id": types.StringType,
	//
	//	"aws_secret_access_key": types.StringType,
	//
	//	"s3_bucket_name": types.StringType,
	//
	//	"s3_bucket_region": types.StringType,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *RedshiftConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
			Configuration: polytomic.RedshiftConnectionConfiguration{
				Hostname:           data.Configuration.Attributes()["hostname"].(types.String).ValueString(),
				Username:           data.Configuration.Attributes()["username"].(types.String).ValueString(),
				Password:           data.Configuration.Attributes()["password"].(types.String).ValueString(),
				Database:           data.Configuration.Attributes()["database"].(types.String).ValueString(),
				Port:               int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
				SSH:                data.Configuration.Attributes()["ssh"].(types.Bool).ValueBool(),
				SSHUser:            data.Configuration.Attributes()["ssh_user"].(types.String).ValueString(),
				SSHHost:            data.Configuration.Attributes()["ssh_host"].(types.String).ValueString(),
				SSHPort:            int(data.Configuration.Attributes()["ssh_port"].(types.Int64).ValueInt64()),
				SSHPrivateKey:      data.Configuration.Attributes()["ssh_private_key"].(types.String).ValueString(),
				AwsAccessKeyID:     data.Configuration.Attributes()["aws_access_key_id"].(types.String).ValueString(),
				AwsSecretAccessKey: data.Configuration.Attributes()["aws_secret_access_key"].(types.String).ValueString(),
				S3BucketName:       data.Configuration.Attributes()["s3_bucket_name"].(types.String).ValueString(),
				S3BucketRegion:     data.Configuration.Attributes()["s3_bucket_region"].(types.String).ValueString(),
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

	//var output polytomic.RedshiftConnectionConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(updated.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"hostname": types.StringType,
	//
	//	"username": types.StringType,
	//
	//	"password": types.StringType,
	//
	//	"database": types.StringType,
	//
	//	"port": types.Int64Type,
	//
	//	"ssh": types.BoolType,
	//
	//	"ssh_user": types.StringType,
	//
	//	"ssh_host": types.StringType,
	//
	//	"ssh_port": types.Int64Type,
	//
	//	"ssh_private_key": types.StringType,
	//
	//	"aws_access_key_id": types.StringType,
	//
	//	"aws_secret_access_key": types.StringType,
	//
	//	"s3_bucket_name": types.StringType,
	//
	//	"s3_bucket_region": types.StringType,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *RedshiftConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
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

func (r *RedshiftConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *RedshiftConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
