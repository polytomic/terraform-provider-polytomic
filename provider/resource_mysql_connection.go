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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &MysqlConnectionResource{}
var _ resource.ResourceWithImportState = &MysqlConnectionResource{}

func (t *MysqlConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: MySQL Connection",
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
						Sensitive:           false,
					},
					"account": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"passwd": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Sensitive:           true,
					},
					"dbname": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh_user": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh_host": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh_port": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"ssh_private_key": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Sensitive:           true,
					},
					"change_detection": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
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
				MarkdownDescription: "MySQL Connection identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MysqlConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_connection"
}

type MysqlConnectionResource struct {
	client *polytomic.Client
}

func (r *MysqlConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:           data.Name.ValueString(),
			Type:           polytomic.MysqlConnectionType,
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.MysqlConnectionConfiguration{
				Hostname:        data.Configuration.Attributes()["hostname"].(types.String).ValueString(),
				Account:         data.Configuration.Attributes()["account"].(types.String).ValueString(),
				Passwd:          data.Configuration.Attributes()["passwd"].(types.String).ValueString(),
				Dbname:          data.Configuration.Attributes()["dbname"].(types.String).ValueString(),
				Port:            int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
				SSH:             data.Configuration.Attributes()["ssh"].(types.Bool).ValueBool(),
				SSHUser:         data.Configuration.Attributes()["ssh_user"].(types.String).ValueString(),
				SSHHost:         data.Configuration.Attributes()["ssh_host"].(types.String).ValueString(),
				SSHPort:         int(data.Configuration.Attributes()["ssh_port"].(types.Int64).ValueInt64()),
				SSHPrivateKey:   data.Configuration.Attributes()["ssh_private_key"].(types.String).ValueString(),
				ChangeDetection: data.Configuration.Attributes()["change_detection"].(types.Bool).ValueBool(),
			},
		},
		polytomic.SkipConfigValidation(),
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating connection: %s", err))
		return
	}
	data.Id = types.StringValue(created.ID)
	data.Name = types.StringValue(created.Name)
	data.Organization = types.StringValue(created.OrganizationId)

	//var output polytomic.MysqlConnectionConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(created.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"hostname": schema.StringAttribute,
	//
	//	"account": schema.StringAttribute,
	//
	//	"passwd": schema.StringAttribute,
	//
	//	"dbname": schema.StringAttribute,
	//
	//	"port": schema.Int64Attribute,
	//
	//	"ssh": schema.BoolAttribute,
	//
	//	"ssh_user": schema.StringAttribute,
	//
	//	"ssh_host": schema.StringAttribute,
	//
	//	"ssh_port": schema.Int64Attribute,
	//
	//	"ssh_private_key": schema.StringAttribute,
	//
	//	"change_detection": schema.BoolAttribute,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Mysql", "id": created.ID})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *MysqlConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	//var output polytomic.MysqlConnectionConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(connection.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"hostname": schema.StringAttribute,
	//
	//	"account": schema.StringAttribute,
	//
	//	"passwd": schema.StringAttribute,
	//
	//	"dbname": schema.StringAttribute,
	//
	//	"port": schema.Int64Attribute,
	//
	//	"ssh": schema.BoolAttribute,
	//
	//	"ssh_user": schema.StringAttribute,
	//
	//	"ssh_host": schema.StringAttribute,
	//
	//	"ssh_port": schema.Int64Attribute,
	//
	//	"ssh_private_key": schema.StringAttribute,
	//
	//	"change_detection": schema.BoolAttribute,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *MysqlConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
			Configuration: polytomic.MysqlConnectionConfiguration{
				Hostname:        data.Configuration.Attributes()["hostname"].(types.String).ValueString(),
				Account:         data.Configuration.Attributes()["account"].(types.String).ValueString(),
				Passwd:          data.Configuration.Attributes()["passwd"].(types.String).ValueString(),
				Dbname:          data.Configuration.Attributes()["dbname"].(types.String).ValueString(),
				Port:            int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
				SSH:             data.Configuration.Attributes()["ssh"].(types.Bool).ValueBool(),
				SSHUser:         data.Configuration.Attributes()["ssh_user"].(types.String).ValueString(),
				SSHHost:         data.Configuration.Attributes()["ssh_host"].(types.String).ValueString(),
				SSHPort:         int(data.Configuration.Attributes()["ssh_port"].(types.Int64).ValueInt64()),
				SSHPrivateKey:   data.Configuration.Attributes()["ssh_private_key"].(types.String).ValueString(),
				ChangeDetection: data.Configuration.Attributes()["change_detection"].(types.Bool).ValueBool(),
			},
		},
		polytomic.SkipConfigValidation(),
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating connection: %s", err))
		return
	}

	data.Id = types.StringValue(updated.ID)
	data.Name = types.StringValue(updated.Name)
	data.Organization = types.StringValue(updated.OrganizationId)

	//var output polytomic.MysqlConnectionConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(updated.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"hostname": schema.StringAttribute,
	//
	//	"account": schema.StringAttribute,
	//
	//	"passwd": schema.StringAttribute,
	//
	//	"dbname": schema.StringAttribute,
	//
	//	"port": schema.Int64Attribute,
	//
	//	"ssh": schema.BoolAttribute,
	//
	//	"ssh_user": schema.StringAttribute,
	//
	//	"ssh_host": schema.StringAttribute,
	//
	//	"ssh_port": schema.Int64Attribute,
	//
	//	"ssh_private_key": schema.StringAttribute,
	//
	//	"change_detection": schema.BoolAttribute,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *MysqlConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *MysqlConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MysqlConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
