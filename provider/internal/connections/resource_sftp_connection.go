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
var _ resource.Resource = &SftpConnectionResource{}
var _ resource.ResourceWithImportState = &SftpConnectionResource{}

var SftpSchema = schema.Schema{
	MarkdownDescription: ":meta:subcategory:Connections: SFTP Connection",
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
				"is_single_table": schema.BoolAttribute{
					MarkdownDescription: `Files are time-based snapshots

    Treat the files as a single table.`,
					Required:  false,
					Optional:  true,
					Computed:  true,
					Sensitive: false,
				},
				"path": schema.StringAttribute{
					MarkdownDescription: `The path to the directory on the SFTP server containing the files.`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"single_table_name": schema.StringAttribute{
					MarkdownDescription: `Collection name`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"skip_lines": schema.Int64Attribute{
					MarkdownDescription: `Skip first lines

    Skip first N lines of each CSV file.`,
					Required:  false,
					Optional:  true,
					Computed:  true,
					Sensitive: false,
				},
				"ssh_host": schema.StringAttribute{
					MarkdownDescription: `Host`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"ssh_port": schema.Int64Attribute{
					MarkdownDescription: `Port`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           false,
				},
				"ssh_private_key": schema.StringAttribute{
					MarkdownDescription: `Private key`,
					Required:            false,
					Optional:            true,
					Computed:            true,
					Sensitive:           true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"ssh_user": schema.StringAttribute{
					MarkdownDescription: `User`,
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
			MarkdownDescription: "SFTP Connection identifier",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	},
}

func (t *SftpConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SftpSchema
}

type SftpConf struct {
	Is_single_table   bool   `mapstructure:"is_single_table" tfsdk:"is_single_table"`
	Path              string `mapstructure:"path" tfsdk:"path"`
	Single_table_name string `mapstructure:"single_table_name" tfsdk:"single_table_name"`
	Skip_lines        int64  `mapstructure:"skip_lines" tfsdk:"skip_lines"`
	Ssh_host          string `mapstructure:"ssh_host" tfsdk:"ssh_host"`
	Ssh_port          int64  `mapstructure:"ssh_port" tfsdk:"ssh_port"`
	Ssh_private_key   string `mapstructure:"ssh_private_key" tfsdk:"ssh_private_key"`
	Ssh_user          string `mapstructure:"ssh_user" tfsdk:"ssh_user"`
}

type SftpConnectionResource struct {
	provider *providerclient.Provider
}

func (r *SftpConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *SftpConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sftp_connection"
}

func (r *SftpConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
		Type:           "sftp",
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

	conf := SftpConf{}
	err = mapstructure.Decode(created.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"is_single_table":   types.BoolType,
		"path":              types.StringType,
		"single_table_name": types.StringType,
		"skip_lines":        types.NumberType,
		"ssh_host":          types.StringType,
		"ssh_port":          types.NumberType,
		"ssh_private_key":   types.StringType,
		"ssh_user":          types.StringType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Sftp", "id": created.Data.Id})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *SftpConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	configAttributes, ok := getConfigAttributes(SftpSchema)
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

	conf := SftpConf{}
	err = mapstructure.Decode(connection.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"is_single_table":   types.BoolType,
		"path":              types.StringType,
		"single_table_name": types.StringType,
		"skip_lines":        types.NumberType,
		"ssh_host":          types.StringType,
		"ssh_port":          types.NumberType,
		"ssh_private_key":   types.StringType,
		"ssh_user":          types.StringType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *SftpConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	configAttributes, ok := getConfigAttributes(SftpSchema)
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

	conf := SftpConf{}
	err = mapstructure.Decode(updated.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(providerclient.ErrorSummary, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"is_single_table":   types.BoolType,
		"path":              types.StringType,
		"single_table_name": types.StringType,
		"skip_lines":        types.NumberType,
		"ssh_host":          types.StringType,
		"ssh_port":          types.NumberType,
		"ssh_private_key":   types.StringType,
		"ssh_user":          types.StringType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *SftpConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *SftpConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
