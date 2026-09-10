package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
var _ resource.Resource = &connectionSchemaFieldResource{}
var _ resource.ResourceWithImportState = &connectionSchemaFieldResource{}

// schemaFieldTypes are the types a user-defined field or override may take.
var schemaFieldTypes = []string{"string", "number", "boolean", "datetime", "array", "object", "binary"}

func NewConnectionSchemaFieldResource() resource.Resource {
	return &connectionSchemaFieldResource{}
}

type connectionSchemaFieldResource struct {
	provider *providerclient.Provider
}

type connectionSchemaFieldResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Organization types.String `tfsdk:"organization"`
	ConnectionID types.String `tfsdk:"connection_id"`
	SchemaID     types.String `tfsdk:"schema_id"`
	FieldID      types.String `tfsdk:"field_id"`
	Label        types.String `tfsdk:"label"`
	Type         types.String `tfsdk:"type"`
	Path         types.String `tfsdk:"path"`
}

func (r *connectionSchemaFieldResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schema_field"
}

func (r *connectionSchemaFieldResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	fieldTypes := make([]string, len(schemaFieldTypes))
	for i, t := range schemaFieldTypes {
		fieldTypes[i] = "`" + t + "`"
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Connection Schema Field\n\n" +
			"Adds a field to a connection schema, or overrides the label, type, or path of a field the source already reports. " +
			"Available on connections that support user-defined fields, such as MongoDB, DynamoDB, Stripe, and file storage connections.\n\n" +
			"Deleting this resource removes an added field, or reverts an overridden field to its detected definition. " +
			"Removing `label`, `type`, or `path` from the configuration keeps the last applied value.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier in the format: organization/connection_id/schema_id/field_id",
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
			"field_id": schema.StringAttribute{
				MarkdownDescription: "ID of the field to add, or of the detected field to override",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Field label. Required when adding a field; defaults to the detected label when overriding one.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Field type, one of %s. Required when adding a field; defaults to the detected type when overriding one.",
					strings.Join(fieldTypes, ", ")),
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(schemaFieldTypes...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "JSONPath used to extract the field's value from each source record, such as `$.address.city`",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *connectionSchemaFieldResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		r.provider = provider
	}
}

func (r *connectionSchemaFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionSchemaFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(ctx, data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	connectionID := data.ConnectionID.ValueString()
	schemaID := data.SchemaID.ValueString()
	fieldID := data.FieldID.ValueString()

	schemaData, err := fetchSchema(ctx, client, connectionID, schemaID)
	if err != nil {
		addSchemaReadError(&resp.Diagnostics, err, connectionID, schemaID)
		return
	}

	var field *polytomic.SchemaField
	switch existing := findSchemaField(schemaData, fieldID); {
	case existing == nil:
		if knownStringPointer(data.Label) == nil || knownStringPointer(data.Type) == nil {
			resp.Diagnostics.AddError(
				"Missing label or type",
				fmt.Sprintf("Schema %s has no field %s, so adding it requires both label and type.", schemaID, fieldID),
			)
			return
		}
		err = client.Schemas.UpsertField(ctx, connectionID, schemaID, &polytomic.UpsertSchemaFieldRequest{
			Fields: []*polytomic.UserFieldRequest{{
				FieldID: fieldID,
				Label:   data.Label.ValueString(),
				Type:    data.Type.ValueString(),
				Path:    knownStringPointer(data.Path),
			}},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error adding field", err.Error())
			return
		}
		// Adding a field returns no body, so read the merged field back.
		schemaData, err = fetchSchema(ctx, client, connectionID, schemaID)
		if err != nil {
			addSchemaReadError(&resp.Diagnostics, err, connectionID, schemaID)
			return
		}
		if field = findSchemaField(schemaData, fieldID); field == nil {
			resp.Diagnostics.AddError(
				"Error adding field",
				fmt.Sprintf("Field %s was not found in schema %s after it was added.", fieldID, schemaID),
			)
			return
		}
	case pointer.GetBool(existing.UserManaged):
		resp.Diagnostics.AddError(
			"Field already has a user-defined definition",
			fmt.Sprintf("Field %s in schema %s is already user-defined or overridden. Import it with the ID %s/%s/%s/%s instead.",
				fieldID, schemaID, orgOrDefault(data.Organization), connectionID, schemaID, fieldID),
		)
		return
	default:
		patch := &polytomic.PatchSchemaFieldRequest{
			Label: knownStringPointer(data.Label),
			Type:  knownStringPointer(data.Type),
			Path:  knownStringPointer(data.Path),
		}
		if patch.Label == nil && patch.Type == nil && patch.Path == nil {
			resp.Diagnostics.AddError(
				"Nothing to override",
				fmt.Sprintf("Set at least one of label, type, or path to override field %s.", fieldID),
			)
			return
		}
		field, err = patchSchemaField(ctx, client, connectionID, schemaID, fieldID, patch)
		if err != nil {
			resp.Diagnostics.AddError("Error overriding field", err.Error())
			return
		}
	}

	applySchemaField(&data, field)

	if data.Organization.IsNull() || data.Organization.IsUnknown() {
		data.Organization = types.StringValue(connectionOrganization(ctx, client, connectionID))
	}
	data.Organization = types.StringValue(orgOrDefault(data.Organization))
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s/%s", data.Organization.ValueString(), connectionID, schemaID, fieldID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemaFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionSchemaFieldResourceModel
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

	field := findSchemaField(schemaData, data.FieldID.ValueString())
	if field == nil || !pointer.GetBool(field.UserManaged) {
		resp.State.RemoveResource(ctx)
		return
	}
	applySchemaField(&data, field)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemaFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state connectionSchemaFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patch := &polytomic.PatchSchemaFieldRequest{}
	if !plan.Label.Equal(state.Label) {
		patch.Label = knownStringPointer(plan.Label)
	}
	if !plan.Type.Equal(state.Type) {
		patch.Type = knownStringPointer(plan.Type)
	}
	if !plan.Path.Equal(state.Path) {
		patch.Path = knownStringPointer(plan.Path)
	}

	if patch.Label != nil || patch.Type != nil || patch.Path != nil {
		client, err := r.provider.Client(ctx, plan.Organization.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error getting client", err.Error())
			return
		}
		field, err := patchSchemaField(ctx, client, plan.ConnectionID.ValueString(), plan.SchemaID.ValueString(), plan.FieldID.ValueString(), patch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating field", err.Error())
			return
		}
		applySchemaField(&plan, field)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionSchemaFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionSchemaFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.provider.Client(ctx, data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	err = client.Schemas.DeleteField(ctx, data.ConnectionID.ValueString(), data.SchemaID.ValueString(), data.FieldID.ValueString())
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error deleting field", err.Error())
	}
}

func (r *connectionSchemaFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	org, connectionID, schemaID, fieldID, err := parseSchemaFieldID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization"), org)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("connection_id"), connectionID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schema_id"), schemaID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("field_id"), fieldID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// parseSchemaFieldID splits an organization/connection_id/schema_id/field_id
// identifier. Schema IDs may themselves contain slashes.
func parseSchemaFieldID(id string) (org, connectionID, schemaID, fieldID string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) >= 4 {
		org, connectionID, fieldID = parts[0], parts[1], parts[len(parts)-1]
		schemaID = strings.Join(parts[2:len(parts)-1], "/")
	}
	if org == "" || connectionID == "" || schemaID == "" || fieldID == "" {
		return "", "", "", "", fmt.Errorf("expected import ID in format: organization/connection_id/schema_id/field_id, got: %s", id)
	}
	return org, connectionID, schemaID, fieldID, nil
}

func patchSchemaField(ctx context.Context, client *ptclient.Client, connectionID, schemaID, fieldID string, patch *polytomic.PatchSchemaFieldRequest) (*polytomic.SchemaField, error) {
	resp, err := client.Schemas.PatchField(ctx, connectionID, schemaID, fieldID, patch)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, errors.New("API returned nil field data")
	}
	return resp.Data, nil
}

func applySchemaField(data *connectionSchemaFieldResourceModel, f *polytomic.SchemaField) {
	data.Label = types.StringPointerValue(f.Name)
	data.Type = types.StringNull()
	if f.Type != nil {
		data.Type = types.StringValue(string(*f.Type))
	}
	data.Path = types.StringPointerValue(f.Path)
}

func addSchemaReadError(diags interface{ AddError(string, string) }, err error, connectionID, schemaID string) {
	if errors.Is(err, errSchemaNotFound) {
		diags.AddError("Schema not found", fmt.Sprintf("Connection %s or schema %s not found", connectionID, schemaID))
		return
	}
	diags.AddError("Error reading schema", err.Error())
}

// knownStringPointer returns nil for a null or unknown value.
func knownStringPointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return pointer.To(v.ValueString())
}

func orgOrDefault(v types.String) string {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return "default"
	}
	return v.ValueString()
}
