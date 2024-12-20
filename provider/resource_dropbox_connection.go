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

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &DropboxConnectionResource{}
var _ resource.ResourceWithImportState = &DropboxConnectionResource{}

func (t *DropboxConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Dropbox Connection",
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
					"app_key": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Default: stringdefault.StaticString(""),
					},
					"app_secret": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Default: stringdefault.StaticString(""),
					},
					"bucket": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"is_single_table": schema.BoolAttribute{
						MarkdownDescription: "Treat the files as a single table.",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"oauth_refresh_token": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Default: stringdefault.StaticString(""),
					},
					"oauth_token_expiry": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Default:             stringdefault.StaticString(""),
					},
					"single_table_file_format": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Default:             stringdefault.StaticString(""),
					},
					"single_table_name": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Default:             stringdefault.StaticString(""),
					},
					"skip_lines": schema.Int64Attribute{
						MarkdownDescription: "Skip first N lines of each CSV file.",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
						Default:             int64default.StaticInt64(0),
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
				MarkdownDescription: "Dropbox Connection identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

type DropboxConf struct {
	App_key string `mapstructure:"app_key" tfsdk:"app_key"`

	App_secret string `mapstructure:"app_secret" tfsdk:"app_secret"`

	Bucket string `mapstructure:"bucket" tfsdk:"bucket"`

	Is_single_table bool `mapstructure:"is_single_table" tfsdk:"is_single_table"`

	Oauth_refresh_token string `mapstructure:"oauth_refresh_token" tfsdk:"oauth_refresh_token"`

	Oauth_token_expiry string `mapstructure:"oauth_token_expiry" tfsdk:"oauth_token_expiry"`

	Single_table_file_format string `mapstructure:"single_table_file_format" tfsdk:"single_table_file_format"`

	Single_table_name string `mapstructure:"single_table_name" tfsdk:"single_table_name"`

	Skip_lines int64 `mapstructure:"skip_lines" tfsdk:"skip_lines"`
}

type DropboxConnectionResource struct {
	provider *client.Provider
}

func (r *DropboxConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *DropboxConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dropbox_connection"
}

func (r *DropboxConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
		Type:           "dropbox",
		OrganizationId: data.Organization.ValueStringPointer(),
		Configuration: map[string]interface{}{
			"app_key":                  data.Configuration.Attributes()["app_key"].(types.String).ValueString(),
			"app_secret":               data.Configuration.Attributes()["app_secret"].(types.String).ValueString(),
			"bucket":                   data.Configuration.Attributes()["bucket"].(types.String).ValueString(),
			"is_single_table":          data.Configuration.Attributes()["is_single_table"].(types.Bool).ValueBool(),
			"oauth_refresh_token":      data.Configuration.Attributes()["oauth_refresh_token"].(types.String).ValueString(),
			"oauth_token_expiry":       data.Configuration.Attributes()["oauth_token_expiry"].(types.String).ValueString(),
			"single_table_file_format": data.Configuration.Attributes()["single_table_file_format"].(types.String).ValueString(),
			"single_table_name":        data.Configuration.Attributes()["single_table_name"].(types.String).ValueString(),
			"skip_lines":               int(data.Configuration.Attributes()["skip_lines"].(types.Int64).ValueInt64()),
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

	conf := DropboxConf{}
	err = mapstructure.Decode(created.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"app_key":                  types.StringType,
		"app_secret":               types.StringType,
		"bucket":                   types.StringType,
		"is_single_table":          types.BoolType,
		"oauth_refresh_token":      types.StringType,
		"oauth_token_expiry":       types.StringType,
		"single_table_file_format": types.StringType,
		"single_table_name":        types.StringType,
		"skip_lines":               types.NumberType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Dropbox", "id": created.Data.Id})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *DropboxConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	conf := DropboxConf{}
	err = mapstructure.Decode(connection.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"app_key":                  types.StringType,
		"app_secret":               types.StringType,
		"bucket":                   types.StringType,
		"is_single_table":          types.BoolType,
		"oauth_refresh_token":      types.StringType,
		"oauth_token_expiry":       types.StringType,
		"single_table_file_format": types.StringType,
		"single_table_name":        types.StringType,
		"skip_lines":               types.NumberType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *DropboxConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
				"app_key":                  data.Configuration.Attributes()["app_key"].(types.String).ValueString(),
				"app_secret":               data.Configuration.Attributes()["app_secret"].(types.String).ValueString(),
				"bucket":                   data.Configuration.Attributes()["bucket"].(types.String).ValueString(),
				"is_single_table":          data.Configuration.Attributes()["is_single_table"].(types.Bool).ValueBool(),
				"oauth_refresh_token":      data.Configuration.Attributes()["oauth_refresh_token"].(types.String).ValueString(),
				"oauth_token_expiry":       data.Configuration.Attributes()["oauth_token_expiry"].(types.String).ValueString(),
				"single_table_file_format": data.Configuration.Attributes()["single_table_file_format"].(types.String).ValueString(),
				"single_table_name":        data.Configuration.Attributes()["single_table_name"].(types.String).ValueString(),
				"skip_lines":               int(data.Configuration.Attributes()["skip_lines"].(types.Int64).ValueInt64()),
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

	conf := DropboxConf{}
	err = mapstructure.Decode(updated.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"app_key":                  types.StringType,
		"app_secret":               types.StringType,
		"bucket":                   types.StringType,
		"is_single_table":          types.BoolType,
		"oauth_refresh_token":      types.StringType,
		"oauth_token_expiry":       types.StringType,
		"single_table_file_format": types.StringType,
		"single_table_name":        types.StringType,
		"skip_lines":               types.NumberType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *DropboxConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *DropboxConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
