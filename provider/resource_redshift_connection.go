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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &RedshiftConnectionResource{}
var _ resource.ResourceWithImportState = &RedshiftConnectionResource{}

func (t *RedshiftConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Redshift Connection",
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
					"hostname": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           true,
					},
					"database": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Default:             int64default.StaticInt64(0),
					},
					"ssh": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"ssh_user": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Default:             stringdefault.StaticString(""),
					},
					"ssh_host": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Default:             stringdefault.StaticString(""),
					},
					"ssh_port": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Default:             int64default.StaticInt64(0),
					},
					"ssh_private_key": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           true,
						Default:             stringdefault.StaticString(""),
					},
					"aws_access_key_id": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Default:             stringdefault.StaticString(""),
					},
					"aws_secret_access_key": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           true,
						Default:             stringdefault.StaticString(""),
					},
					"s3_bucket_name": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Default:             stringdefault.StaticString(""),
					},
					"s3_bucket_region": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Default:             stringdefault.StaticString(""),
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
				MarkdownDescription: "Redshift Connection identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
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
		polytomic.WithIdempotencyKey(uuid.NewString()),
		polytomic.SkipConfigValidation(),
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}
	data.Id = types.StringValue(created.ID)
	data.Name = types.StringValue(created.Name)
	data.Organization = types.StringValue(created.OrganizationId)

	var output polytomic.RedshiftConnectionConfiguration
	cfg := &mapstructure.DecoderConfig{
		Result: &output,
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	decoder.Decode(created.Configuration)
	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"hostname":              types.StringType,
		"username":              types.StringType,
		"password":              types.StringType,
		"database":              types.StringType,
		"port":                  types.NumberType,
		"ssh":                   types.BoolType,
		"ssh_user":              types.StringType,
		"ssh_host":              types.StringType,
		"ssh_port":              types.NumberType,
		"ssh_private_key":       types.StringType,
		"aws_access_key_id":     types.StringType,
		"aws_secret_access_key": types.StringType,
		"s3_bucket_name":        types.StringType,
		"s3_bucket_region":      types.StringType,
	}, output)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

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

	var output polytomic.RedshiftConnectionConfiguration
	cfg := &mapstructure.DecoderConfig{
		Result: &output,
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	decoder.Decode(connection.Configuration)
	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"hostname":              types.StringType,
		"username":              types.StringType,
		"password":              types.StringType,
		"database":              types.StringType,
		"port":                  types.NumberType,
		"ssh":                   types.BoolType,
		"ssh_user":              types.StringType,
		"ssh_host":              types.StringType,
		"ssh_port":              types.NumberType,
		"ssh_private_key":       types.StringType,
		"aws_access_key_id":     types.StringType,
		"aws_secret_access_key": types.StringType,
		"s3_bucket_name":        types.StringType,
		"s3_bucket_region":      types.StringType,
	}, output)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

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
		polytomic.WithIdempotencyKey(uuid.NewString()),
		polytomic.SkipConfigValidation(),
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	data.Id = types.StringValue(updated.ID)
	data.Name = types.StringValue(updated.Name)
	data.Organization = types.StringValue(updated.OrganizationId)

	var output polytomic.RedshiftConnectionConfiguration
	cfg := &mapstructure.DecoderConfig{
		Result: &output,
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	decoder.Decode(updated.Configuration)
	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"hostname":              types.StringType,
		"username":              types.StringType,
		"password":              types.StringType,
		"database":              types.StringType,
		"port":                  types.NumberType,
		"ssh":                   types.BoolType,
		"ssh_user":              types.StringType,
		"ssh_host":              types.StringType,
		"ssh_port":              types.NumberType,
		"ssh_private_key":       types.StringType,
		"aws_access_key_id":     types.StringType,
		"aws_secret_access_key": types.StringType,
		"s3_bucket_name":        types.StringType,
		"s3_bucket_region":      types.StringType,
	}, output)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

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

	if data.ForceDestroy.ValueBool() {
		err := r.client.Connections().Delete(ctx, uuid.MustParse(data.Id.ValueString()), polytomic.WithForceDelete())
		if err != nil {
			pErr := polytomic.ApiError{}
			if errors.As(err, &pErr) {
				if pErr.StatusCode == http.StatusNotFound {
					resp.State.RemoveResource(ctx)
					return
				}
			}
			resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting connection: %s", err))
		}
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
