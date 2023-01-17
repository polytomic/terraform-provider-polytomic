// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package provider

import (
	"context"
	"fmt"

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
var _ resource.Resource = &MongodbConnectionResource{}
var _ resource.ResourceWithImportState = &MongodbConnectionResource{}

func (t *MongodbConnectionResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: MongoDB Connection",
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
					"hosts": {
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
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"srv": {
						MarkdownDescription: "",
						Type:                types.BoolType,
						Required:            false,
						Optional:            true,
						Sensitive:           false,
					},
					"params": {
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
				MarkdownDescription: "MongoDB Connection identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (r *MongodbConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mongodb_connection"
}

type MongodbConnectionResource struct {
	client *polytomic.Client
}

func (r *MongodbConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:           data.Name.ValueString(),
			Type:           polytomic.MongoDBConnectionType,
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.MongoDBConfiguration{
				Hosts:    data.Configuration.Attributes()["hosts"].(types.String).ValueString(),
				Username: data.Configuration.Attributes()["username"].(types.String).ValueString(),
				Password: data.Configuration.Attributes()["password"].(types.String).ValueString(),
				Database: data.Configuration.Attributes()["database"].(types.String).ValueString(),
				SRV:      data.Configuration.Attributes()["srv"].(types.Bool).ValueBool(),
				Params:   data.Configuration.Attributes()["params"].(types.String).ValueString(),
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

	//var output polytomic.MongoDBConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(created.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"hosts": types.StringType,
	//
	//	"username": types.StringType,
	//
	//	"password": types.StringType,
	//
	//	"database": types.StringType,
	//
	//	"srv": types.BoolType,
	//
	//	"params": types.StringType,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Mongodb", "id": created.ID})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *MongodbConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	connection, err := r.client.Connections().Get(ctx, uuid.MustParse(data.Id.ValueString()))
	if err != nil {
		if err.Error() == ConnectionNotFoundErr {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading connection: %s", err))
		return
	}

	data.Id = types.StringValue(connection.ID)
	data.Name = types.StringValue(connection.Name)
	data.Organization = types.StringValue(connection.OrganizationId)

	//var output polytomic.MongoDBConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(connection.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"hosts": types.StringType,
	//
	//	"username": types.StringType,
	//
	//	"password": types.StringType,
	//
	//	"database": types.StringType,
	//
	//	"srv": types.BoolType,
	//
	//	"params": types.StringType,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *MongodbConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
			Configuration: polytomic.MongoDBConfiguration{
				Hosts:    data.Configuration.Attributes()["hosts"].(types.String).ValueString(),
				Username: data.Configuration.Attributes()["username"].(types.String).ValueString(),
				Password: data.Configuration.Attributes()["password"].(types.String).ValueString(),
				Database: data.Configuration.Attributes()["database"].(types.String).ValueString(),
				SRV:      data.Configuration.Attributes()["srv"].(types.Bool).ValueBool(),
				Params:   data.Configuration.Attributes()["params"].(types.String).ValueString(),
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

	//var output polytomic.MongoDBConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(updated.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"hosts": types.StringType,
	//
	//	"username": types.StringType,
	//
	//	"password": types.StringType,
	//
	//	"database": types.StringType,
	//
	//	"srv": types.BoolType,
	//
	//	"params": types.StringType,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *MongodbConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Connections().Delete(ctx, uuid.MustParse(data.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting connection: %s", err))
		return
	}
}

func (r *MongodbConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MongodbConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
