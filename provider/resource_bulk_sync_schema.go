package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &bulkSyncSchemaResource{}
var _ resource.ResourceWithImportState = &bulkSyncSchemaResource{}

func (r *bulkSyncSchemaResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"sync_id": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"partition_key": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
			},
		},
	}, nil
}

func (r *bulkSyncSchemaResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || !req.State.Raw.IsKnown() {
		return
	}

	config := &bulkSyncSchemaResourceData{}
	resp.Diagnostics.Append(req.Config.Get(ctx, config)...)

	plan := &bulkSyncSchemaResourceData{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.PartitionKey.IsNull() {
		plan.PartitionKey = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)

}

func (r *bulkSyncSchemaResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bulkSyncSchemaResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_sync_schema"
}

type bulkSyncSchemaResourceData struct {
	Id           types.String `tfsdk:"id"`
	SyncID       types.String `tfsdk:"sync_id"`
	PartitionKey types.String `tfsdk:"partition_key"`
}

type bulkSyncSchemaResource struct {
	client *polytomic.Client
}

func (r *bulkSyncSchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data bulkSyncSchemaResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	existingSchemas, err := r.client.Bulk().GetBulkSyncSchemas(ctx, data.SyncID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error getting bulk sync schemas: %s", err))
		return
	}

	if len(existingSchemas) > 0 {
		for _, existingSchema := range existingSchemas {
			if existingSchema.ID == data.Id.ValueString() && existingSchema.Enabled {
				resp.Diagnostics.AddError(clientError, fmt.Sprintf("Bulk sync schema already exists: %s", data.Id.ValueString()))
				return
			}
		}
	}
	newSchemas := append(existingSchemas, polytomic.BulkSchema{
		ID:           data.Id.ValueString(),
		Enabled:      true,
		PartitionKey: data.PartitionKey.ValueString(),
	})

	_, err = r.client.Bulk().UpdateBulkSyncSchemas(ctx, data.SyncID.ValueString(), newSchemas)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error creating bulk sync: %s", err))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bulkSyncSchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data bulkSyncSchemaResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	schema, err := r.client.Bulk().GetBulkSyncSchema(ctx, data.SyncID.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error reading bulk sync: %s", err))
		return
	}

	data.Id = types.StringValue(schema.ID)
	data.PartitionKey = types.StringValue(schema.PartitionKey)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

}

func (r *bulkSyncSchemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data bulkSyncSchemaResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.Bulk().UpdateBulkSyncSchema(ctx,
		data.SyncID.ValueString(),
		data.Id.ValueString(),
		polytomic.BulkSchema{
			Enabled:      true,
			PartitionKey: data.PartitionKey.ValueString(),
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error updating organization: %s", err))
		return
	}

	data.PartitionKey = types.StringValue(updated.PartitionKey)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)

	return
}

func (r *bulkSyncSchemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data bulkSyncSchemaResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	existingSchemas, err := r.client.Bulk().GetBulkSyncSchemas(ctx, data.SyncID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error getting bulk sync schemas: %s", err))
		return
	}

	var schemas []polytomic.BulkSchema
	if len(existingSchemas) > 0 {
		for _, existingSchema := range existingSchemas {
			if existingSchema.ID == data.Id.ValueString() {
				existingSchema.Enabled = false
			}
			schemas = append(schemas, existingSchema)

		}
	}

	_, err = r.client.Bulk().UpdateBulkSyncSchemas(ctx, data.SyncID.ValueString(), schemas)
	if err != nil {
		resp.Diagnostics.AddError(clientError, fmt.Sprintf("Error deleting bulk sync schema: %s", err))
		return
	}

}

func (r *bulkSyncSchemaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ".")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: sync_id.schema_id. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("sync_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
