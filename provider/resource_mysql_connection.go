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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
	ptcore "github.com/polytomic/polytomic-go/core"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/client"
)

const (
	// ExplicitSuppress strategy suppresses "(known after changes)" messages unless we're in the initial creation
	ExplicitSuppress = iota
)

// SuppressDiff returns a plan modifier that propagates a state value into the planned value, when it is Known, and the Plan Value is Unknown
func SuppressDiff(strategy int) planmodifier.Object {
	return suppressDiff{
		strategy: strategy,
	}
}

// suppressDiff implements the plan modifier.
type suppressDiff struct {
	strategy int
}

// Description returns a human-readable description of the plan modifier.
func (m suppressDiff) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m suppressDiff) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// PlanModifyObject implements the plan modification logic.
func (m suppressDiff) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value
	if req.ConfigValue.IsUnknown() {
		return
	}

	if IsAllStateUnknown(ctx, req.State) {
		return
	}

	resp.PlanValue = req.StateValue
}

func IsAllStateUnknown(ctx context.Context, state tfsdk.State) bool {
	attrs := state.Schema.GetAttributes()
	anyFound := false
	for k, _ := range attrs {
		attrValue := new(attr.Value)
		state.GetAttribute(ctx, path.Root(k), attrValue)
		if attrValue != nil && !(*attrValue).IsUnknown() && !(*attrValue).IsNull() {
			anyFound = true
			break
		}
	}

	return !anyFound
}

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
					"account": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"change_detection": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"dbname": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            true,
						Sensitive:           false,
					},
					"hostname": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"passwd": schema.StringAttribute{
						MarkdownDescription: "",
						Optional:            true,
						Computed:            true,
						Sensitive:           true,
					},
					"port": schema.Int64Attribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Computed:            false,
						Sensitive:           false,
					},
					"ssl": schema.BoolAttribute{
						MarkdownDescription: "",
						Required:            false,
						Optional:            true,
						Computed:            false,
						Sensitive:           false,
					},
				},

				Required: true,
				// PlanModifiers: []planmodifier.Object{
				// 	SuppressDiff(ExplicitSuppress),
				// },
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

type MysqlConf struct {
	Account string `mapstructure:"account" tfsdk:"account"`

	Change_detection bool `mapstructure:"change_detection" tfsdk:"change_detection"`

	Dbname string `mapstructure:"dbname" tfsdk:"dbname"`

	Hostname string `mapstructure:"hostname" tfsdk:"hostname"`

	Passwd string `mapstructure:"passwd" tfsdk:"passwd"`

	Port int64 `mapstructure:"port" tfsdk:"port"`

	Ssl bool `mapstructure:"ssl" tfsdk:"ssl"`
}

type MysqlConnectionResource struct {
	provider *client.Provider
}

func (r *MysqlConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *MysqlConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_connection"
}

func (r *MysqlConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
		Type:           "mysql",
		OrganizationId: data.Organization.ValueStringPointer(),
		Configuration: map[string]interface{}{
			"account":          data.Configuration.Attributes()["account"].(types.String).ValueString(),
			"change_detection": data.Configuration.Attributes()["change_detection"].(types.Bool).ValueBool(),
			"dbname":           data.Configuration.Attributes()["dbname"].(types.String).ValueString(),
			"hostname":         data.Configuration.Attributes()["hostname"].(types.String).ValueString(),
			"passwd":           data.Configuration.Attributes()["passwd"].(types.String).ValueString(),
			"port":             int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
			"ssl":              data.Configuration.Attributes()["ssl"].(types.Bool).ValueBool(),
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

	conf := MysqlConf{}
	err = mapstructure.Decode(created.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	newVal, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"account":          types.StringType,
		"change_detection": types.BoolType,
		"dbname":           types.StringType,
		"hostname":         types.StringType,
		"passwd":           types.StringType,
		"port":             types.NumberType,
		"ssl":              types.BoolType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	fmt.Printf("\n\n!!!!%#v\n\n", newVal)
	newVal.Attributes()["passwd"] = data.Configuration.Attributes()["passwd"]
	data.Configuration = newVal
	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Mysql", "id": created.Data.Id})

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

	conf := MysqlConf{}
	err = mapstructure.Decode(connection.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"account":          types.StringType,
		"change_detection": types.BoolType,
		"dbname":           types.StringType,
		"hostname":         types.StringType,
		"passwd":           types.StringType,
		"port":             types.NumberType,
		"ssl":              types.BoolType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

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
				"account":          data.Configuration.Attributes()["account"].(types.String).ValueString(),
				"change_detection": data.Configuration.Attributes()["change_detection"].(types.Bool).ValueBool(),
				"dbname":           data.Configuration.Attributes()["dbname"].(types.String).ValueString(),
				"hostname":         data.Configuration.Attributes()["hostname"].(types.String).ValueString(),
				"passwd":           data.Configuration.Attributes()["passwd"].(types.String).ValueString(),
				"port":             int(data.Configuration.Attributes()["port"].(types.Int64).ValueInt64()),
				"ssl":              data.Configuration.Attributes()["ssl"].(types.Bool).ValueBool(),
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

	conf := MysqlConf{}
	err = mapstructure.Decode(updated.Data.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error decoding connection configuration: %s", err))
	}

	data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
		"account":          types.StringType,
		"change_detection": types.BoolType,
		"dbname":           types.StringType,
		"hostname":         types.StringType,
		"passwd":           types.StringType,
		"port":             types.NumberType,
		"ssl":              types.BoolType,
	}, conf)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
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

	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting connection: %s", err))
	}
}

func (r *MysqlConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
