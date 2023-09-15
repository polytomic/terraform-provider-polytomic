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
var _ resource.Resource = &CosmosdbConnectionResource{}
var _ resource.ResourceWithImportState = &CosmosdbConnectionResource{}

func (t *CosmosdbConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Cosmos DB Connection",
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
					"uri": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Sensitive:           false,
					},
					"key": schema.StringAttribute{
						MarkdownDescription: "",
						Required:            true,
						Optional:            false,
						Sensitive:           true,
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
				MarkdownDescription: "Cosmos DB Connection identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *CosmosdbConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cosmosdb_connection"
}

type CosmosdbConnectionResource struct {
	client *polytomic.Client
}

func (r *CosmosdbConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Connections().Create(ctx,
		polytomic.CreateConnectionMutation{
			Name:           data.Name.ValueString(),
			Type:           polytomic.CosmosDBConnectionType,
			OrganizationId: data.Organization.ValueString(),
			Configuration: polytomic.CosmosDBConnectionConfiguration{
				URI: data.Configuration.Attributes()["uri"].(types.String).ValueString(),
				Key: data.Configuration.Attributes()["key"].(types.String).ValueString(),
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

	//var output polytomic.CosmosDBConnectionConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(created.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"uri": schema.StringAttribute,
	//
	//	"key": schema.StringAttribute,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	tflog.Trace(ctx, "created a connection", map[string]interface{}{"type": "Cosmosdb", "id": created.ID})

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *CosmosdbConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	//var output polytomic.CosmosDBConnectionConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(connection.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"uri": schema.StringAttribute,
	//
	//	"key": schema.StringAttribute,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *CosmosdbConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
			Configuration: polytomic.CosmosDBConnectionConfiguration{
				URI: data.Configuration.Attributes()["uri"].(types.String).ValueString(),
				Key: data.Configuration.Attributes()["key"].(types.String).ValueString(),
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

	//var output polytomic.CosmosDBConnectionConfiguration
	//cfg := &mapstructure.DecoderConfig{
	//    Result:   &output,
	//}
	//decoder, _ := mapstructure.NewDecoder(cfg)
	//decoder.Decode(updated.Configuration)
	//data.Configuration, diags = types.ObjectValueFrom(ctx, map[string]attr.Type{
	//
	//	"uri": schema.StringAttribute,
	//
	//	"key": schema.StringAttribute,
	//
	//}, output)
	//if diags.HasError() {
	//	resp.Diagnostics.Append(diags...)
	//	return
	//}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *CosmosdbConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *CosmosdbConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *CosmosdbConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
