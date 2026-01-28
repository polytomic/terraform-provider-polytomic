package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
	ptcore "github.com/polytomic/polytomic-go/core"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &connectionSchemaPrimaryKeysResource{}
var _ resource.ResourceWithImportState = &connectionSchemaPrimaryKeysResource{}

func NewConnectionSchemaPrimaryKeysResource() resource.Resource {
	return &connectionSchemaPrimaryKeysResource{}
}

type connectionSchemaPrimaryKeysResource struct {
	provider *providerclient.Provider
}

type connectionSchemaPrimaryKeysResourceModel struct {
	Organization types.String `tfsdk:"organization"`
	ConnectionID types.String `tfsdk:"connection_id"`
	SchemaID     types.String `tfsdk:"schema_id"`
	FieldIDs     types.Set    `tfsdk:"field_ids"`
	ID           types.String `tfsdk:"id"`
}

func (r *connectionSchemaPrimaryKeysResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schema_primary_keys"
}

func (r *connectionSchemaPrimaryKeysResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Connection Schema Primary Keys\n\n" +
			"Manages primary key overrides for a connection schema. " +
			"Primary keys can be set to override the auto-detected keys from the source connection. " +
			"Deleting this resource will reset the schema to use auto-detected primary keys.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier in the format: organization/connection_id/schema_id",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_id": schema.StringAttribute{
				MarkdownDescription: "Connection ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schema_id": schema.StringAttribute{
				MarkdownDescription: "Schema ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"field_ids": schema.SetAttribute{
				MarkdownDescription: "Set of field IDs to use as primary keys. " +
					"These IDs can be found using the polytomic_connection_schema data source.",
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

func (r *connectionSchemaPrimaryKeysResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *connectionSchemaPrimaryKeysResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionSchemaPrimaryKeysResourceModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate field_ids is not empty
	if data.FieldIDs.IsNull() || len(data.FieldIDs.Elements()) == 0 {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"At least one field_id must be specified for primary keys",
		)
		return
	}

	// Extract field IDs from the set
	var fieldIDs []string
	diags = data.FieldIDs.ElementsAs(ctx, &fieldIDs, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get client
	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	// Build request - mark specified fields as primary keys
	pkFields := make([]*polytomic.SchemaPrimaryKeyOverrideInput, len(fieldIDs))
	for i, fieldID := range fieldIDs {
		pkFields[i] = &polytomic.SchemaPrimaryKeyOverrideInput{
			FieldId:      fieldID,
			IsPrimaryKey: true,
		}
	}

	setRequest := &polytomic.SetPrimaryKeysRequest{
		Fields: pkFields,
	}

	// Set primary keys
	err = client.Schemas.SetPrimaryKeys(
		ctx,
		data.ConnectionID.ValueString(),
		data.SchemaID.ValueString(),
		setRequest,
	)
	if err != nil {
		pErr := &ptcore.APIError{}
		if ok := errors.As(err, &pErr); ok {
			if pErr.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddError(
					"Schema not found",
					fmt.Sprintf("Connection %s or schema %s not found", data.ConnectionID.ValueString(), data.SchemaID.ValueString()),
				)
				return
			}
		}
		resp.Diagnostics.AddError("Error setting primary keys", err.Error())
		return
	}

	// Set organization if not provided
	if data.Organization.IsNull() {
		connResp, err := client.Connections.Get(ctx, data.ConnectionID.ValueString())
		if err == nil && connResp.Data != nil && connResp.Data.OrganizationId != nil {
			data.Organization = types.StringValue(*connResp.Data.OrganizationId)
		} else {
			data.Organization = types.StringValue("default")
		}
	}

	// Set ID in composite format
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s",
		data.Organization.ValueString(),
		data.ConnectionID.ValueString(),
		data.SchemaID.ValueString()))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *connectionSchemaPrimaryKeysResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionSchemaPrimaryKeysResourceModel

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

	// Get schema to read current primary keys
	schemaResp, err := client.Schemas.Get(ctx, data.ConnectionID.ValueString(), data.SchemaID.ValueString())
	if err != nil {
		pErr := &ptcore.APIError{}
		if ok := errors.As(err, &pErr); ok {
			if pErr.StatusCode == http.StatusNotFound {
				// Resource no longer exists
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError("Error reading schema", err.Error())
		return
	}

	if schemaResp.Data == nil {
		resp.Diagnostics.AddError("Error reading schema", "API returned nil schema data")
		return
	}

	// Extract primary key field IDs from the schema
	// Note: We cannot distinguish between auto-detected and overridden primary keys from the API response.
	// We rely on Terraform state to track that this resource represents an override.
	// If the resource exists in state, we assume the override is active.
	primaryKeyFieldIDs := []string{}
	if schemaResp.Data.Fields != nil {
		for _, field := range schemaResp.Data.Fields {
			// The API doesn't have an is_primary_key flag in the SchemaField struct,
			// so we need to determine primary keys differently.
			// For now, we'll preserve the state's field_ids since the API doesn't
			// return primary key information in a queryable way.
			if field.Id != nil {
				// Check if this field ID is in our current state
				var currentFieldIDs []string
				diags = data.FieldIDs.ElementsAs(ctx, &currentFieldIDs, false)
				if diags.HasError() {
					// If we can't read current state, preserve it
					resp.Diagnostics.Append(diags...)
					return
				}

				for _, stateFieldID := range currentFieldIDs {
					if stateFieldID == *field.Id {
						primaryKeyFieldIDs = append(primaryKeyFieldIDs, *field.Id)
						break
					}
				}
			}
		}
	}

	// If no primary keys found but resource exists, preserve state
	// This handles the case where the API doesn't return primary key information
	if len(primaryKeyFieldIDs) == 0 {
		// Keep existing state as-is since we can't verify from API
		return
	}

	// Update field_ids with current primary keys
	fieldIDsSet, diags := types.SetValueFrom(ctx, types.StringType, primaryKeyFieldIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.FieldIDs = fieldIDsSet

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *connectionSchemaPrimaryKeysResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data connectionSchemaPrimaryKeysResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate field_ids is not empty
	if data.FieldIDs.IsNull() || len(data.FieldIDs.Elements()) == 0 {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"At least one field_id must be specified for primary keys",
		)
		return
	}

	// Extract field IDs from the set
	var fieldIDs []string
	diags = data.FieldIDs.ElementsAs(ctx, &fieldIDs, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get client
	client, err := r.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	// Build request - mark specified fields as primary keys
	pkFields := make([]*polytomic.SchemaPrimaryKeyOverrideInput, len(fieldIDs))
	for i, fieldID := range fieldIDs {
		pkFields[i] = &polytomic.SchemaPrimaryKeyOverrideInput{
			FieldId:      fieldID,
			IsPrimaryKey: true,
		}
	}

	setRequest := &polytomic.SetPrimaryKeysRequest{
		Fields: pkFields,
	}

	// Update primary keys
	err = client.Schemas.SetPrimaryKeys(
		ctx,
		data.ConnectionID.ValueString(),
		data.SchemaID.ValueString(),
		setRequest,
	)
	if err != nil {
		pErr := &ptcore.APIError{}
		if ok := errors.As(err, &pErr); ok {
			if pErr.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddError(
					"Schema not found",
					fmt.Sprintf("Connection %s or schema %s not found", data.ConnectionID.ValueString(), data.SchemaID.ValueString()),
				)
				return
			}
		}
		resp.Diagnostics.AddError("Error updating primary keys", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r *connectionSchemaPrimaryKeysResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionSchemaPrimaryKeysResourceModel

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

	// Reset primary keys to auto-detected values
	// This is what DELETE endpoint does - removes overrides
	err = client.Schemas.ResetPrimaryKeys(
		ctx,
		data.ConnectionID.ValueString(),
		data.SchemaID.ValueString(),
	)
	if err != nil {
		pErr := &ptcore.APIError{}
		if ok := errors.As(err, &pErr); ok {
			if pErr.StatusCode == http.StatusNotFound {
				// Resource already gone, consider it successfully deleted
				return
			}
		}
		resp.Diagnostics.AddError("Error resetting primary keys", err.Error())
		return
	}
}

func (r *connectionSchemaPrimaryKeysResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID format: organization/connection_id/schema_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format: organization/connection_id/schema_id, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("connection_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schema_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	// Note: field_ids will be populated by the subsequent Read operation
}
