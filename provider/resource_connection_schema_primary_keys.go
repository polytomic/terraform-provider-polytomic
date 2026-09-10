package provider

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go/v25"
	ptclient "github.com/polytomic/polytomic-go/v25/client"
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
		MarkdownDescription: ":meta:subcategory:Schemas: Connection Schema Primary Keys\n\n" +
			"Sets the primary key of a connection schema, overriding the keys detected from the source connection. " +
			"`field_ids` is the schema's complete primary key: listed fields are marked as keys, and any other field the source reports as a key is unmarked.\n\n" +
			"Deleting this resource removes every primary key override on the schema, restoring the detected keys.",
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
				MarkdownDescription: "IDs of the fields that make up the schema's primary key. " +
					"Fields the source reports as keys are unmarked unless listed here. " +
					"These IDs can be found using the polytomic_connection_schema data source. " +
					"To make a field added with `polytomic_connection_schema_field` a key, reference its `field_id` " +
					"so that Terraform adds the field first.",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
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
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(ctx, data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	setPrimaryKeys(ctx, client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Organization = types.StringValue(resourceOrganization(ctx, client, data.Organization, data.ConnectionID.ValueString()))
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s",
		data.Organization.ValueString(),
		data.ConnectionID.ValueString(),
		data.SchemaID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemaPrimaryKeysResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionSchemaPrimaryKeysResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(ctx, data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	schemaData, err := fetchSchema(ctx, client, data.ConnectionID.ValueString(), data.SchemaID.ValueString())
	if errors.Is(err, errSchemaNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading schema", err.Error())
		return
	}

	var diags diag.Diagnostics
	data.FieldIDs, diags = types.SetValueFrom(ctx, types.StringType, effectivePrimaryKeys(schemaData.Fields))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemaPrimaryKeysResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data connectionSchemaPrimaryKeysResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(ctx, data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	setPrimaryKeys(ctx, client, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemaPrimaryKeysResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionSchemaPrimaryKeysResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(ctx, data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	err = client.Schemas.ResetPrimaryKeys(ctx, data.ConnectionID.ValueString(), data.SchemaID.ValueString())
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error resetting primary keys", err.Error())
	}
}

func (r *connectionSchemaPrimaryKeysResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID format: organization/connection_id/schema_id; schema IDs may contain slashes.
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
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
}

// setPrimaryKeys makes data.FieldIDs the schema's complete primary key.
func setPrimaryKeys(ctx context.Context, client *ptclient.Client, data *connectionSchemaPrimaryKeysResourceModel, diags *diag.Diagnostics) {
	var fieldIDs []string
	diags.Append(data.FieldIDs.ElementsAs(ctx, &fieldIDs, false)...)
	if diags.HasError() {
		return
	}

	connectionID := data.ConnectionID.ValueString()
	schemaID := data.SchemaID.ValueString()
	schemaData, err := fetchSchema(ctx, client, connectionID, schemaID)
	if err != nil {
		addSchemaReadError(diags, err, connectionID, schemaID)
		return
	}

	overrides, err := primaryKeyOverrides(schemaData.Fields, fieldIDs)
	if err != nil {
		diags.AddError("Invalid field_ids", fmt.Sprintf("Schema %s: %s", schemaID, err))
		return
	}

	err = client.Schemas.SetPrimaryKeys(ctx, connectionID, schemaID, &polytomic.SetPrimaryKeysRequest{
		Fields: overrides,
	})
	if err != nil {
		diags.AddError("Error setting primary keys", err.Error())
	}
}

// primaryKeyOverrides returns the overrides that make fieldIDs the complete
// primary key. Every other field that is, was, or could revert to being a key
// is explicitly unmarked, so the result holds whether the API merges overrides
// into existing ones or replaces them.
func primaryKeyOverrides(fields []*polytomic.SchemaField, fieldIDs []string) ([]*polytomic.SchemaPrimaryKeyOverrideInput, error) {
	missing := make(map[string]bool, len(fieldIDs))
	for _, id := range fieldIDs {
		missing[id] = true
	}

	overrides := []*polytomic.SchemaPrimaryKeyOverrideInput{}
	for _, f := range fields {
		if f == nil {
			continue
		}
		id := pointer.GetString(f.ID)
		switch {
		case missing[id]:
			overrides = append(overrides, &polytomic.SchemaPrimaryKeyOverrideInput{FieldID: id, IsPrimaryKey: true})
			delete(missing, id)
		case pointer.GetBool(f.IsPrimaryKey) || pointer.GetBool(f.SourcePrimaryKey) || f.PrimaryKeyOverride != nil:
			overrides = append(overrides, &polytomic.SchemaPrimaryKeyOverrideInput{FieldID: id, IsPrimaryKey: false})
		}
	}

	if len(missing) > 0 {
		ids := make([]string, 0, len(missing))
		for id := range missing {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		return nil, fmt.Errorf("no such fields: %s", strings.Join(ids, ", "))
	}
	return overrides, nil
}

// effectivePrimaryKeys returns the IDs of the fields that currently make up the
// schema's primary key, including overrides.
func effectivePrimaryKeys(fields []*polytomic.SchemaField) []string {
	keys := []string{}
	for _, f := range fields {
		if f != nil && pointer.GetBool(f.IsPrimaryKey) {
			keys = append(keys, pointer.GetString(f.ID))
		}
	}
	sort.Strings(keys)
	return keys
}
