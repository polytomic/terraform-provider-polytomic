package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/polytomic/polytomic-go/v25"
	ptclient "github.com/polytomic/polytomic-go/v25/client"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &connectionSchemaFieldResource{}
var _ resource.ResourceWithImportState = &connectionSchemaFieldResource{}
var _ resource.ResourceWithModifyPlan = &connectionSchemaFieldResource{}
var _ resource.ResourceWithValidateConfig = &connectionSchemaFieldResource{}

func NewConnectionSchemaFieldResource() resource.Resource {
	return &connectionSchemaFieldResource{}
}

type connectionSchemaFieldResource struct {
	provider *providerclient.Provider
}

type connectionSchemaFieldResourceModel struct {
	ID           types.String  `tfsdk:"id"`
	Organization types.String  `tfsdk:"organization"`
	ConnectionID types.String  `tfsdk:"connection_id"`
	SchemaID     types.String  `tfsdk:"schema_id"`
	FieldID      types.String  `tfsdk:"field_id"`
	Label        types.String  `tfsdk:"label"`
	Type         types.String  `tfsdk:"type"`
	Precision    types.Int64   `tfsdk:"precision"`
	Scale        types.Int64   `tfsdk:"scale"`
	TypeSpec     typeSpecValue `tfsdk:"type_spec"`
	Path         types.String  `tfsdk:"path"`
}

func (r *connectionSchemaFieldResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schema_field"
}

func (r *connectionSchemaFieldResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	quoted := func(names []string) string {
		q := make([]string, len(names))
		for i, n := range names {
			q[i] = "`" + n + "`"
		}
		return strings.Join(q, ", ")
	}
	basicCount := len(basicFieldTypes)

	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Connection Schema Field\n\n" +
			"Adds a field to a connection schema, or overrides the label, type, or path of a field the source already reports. " +
			"Available on connections that support user-defined fields, such as MongoDB, DynamoDB, Stripe, and file storage connections.\n\n" +
			"Deleting this resource removes an added field, or reverts an overridden field to its detected definition. " +
			"Removing `label`, `type`, `type_spec`, or `path` from the configuration keeps the last applied value.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier in the format: organization/connection_id/schema_id/field_id, with `%` and `/` in the field ID escaped as `%25` and `%2F`",
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
				MarkdownDescription: fmt.Sprintf("Field type: one of %s, or a detailed type: %s. "+
					"`decimal` also requires `precision` and `scale`. "+
					"Adding a field requires `type` or `type_spec`; overriding one defaults to the detected type.",
					quoted(fieldTypeNames[:basicCount]), quoted(fieldTypeNames[basicCount:])),
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(fieldTypeNames...),
				},
			},
			"precision": schema.Int64Attribute{
				MarkdownDescription: "Total number of digits of a `decimal` field",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"scale": schema.Int64Attribute{
				MarkdownDescription: "Number of digits after the decimal point of a `decimal` field",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"type_spec": schema.StringAttribute{
				MarkdownDescription: "The field's detailed type, JSON encoded, for types `type` cannot express, such as " +
					"`jsonencode([\"array\", \"string\"])`. Always reports the field's current type.",
				CustomType: typeSpecType{},
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("type"), path.MatchRoot("precision"), path.MatchRoot("scale")),
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

func (r *connectionSchemaFieldResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data connectionSchemaFieldResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() || data.Type.IsUnknown() || data.Precision.IsUnknown() || data.Scale.IsUnknown() {
		return
	}

	decimal := data.Type.ValueString() == "decimal"
	switch {
	case decimal && (data.Precision.IsNull() || data.Scale.IsNull()):
		resp.Diagnostics.AddAttributeError(path.Root("type"), "Missing precision or scale",
			"A decimal field requires both precision and scale.")
	case !decimal && (!data.Precision.IsNull() || !data.Scale.IsNull()):
		resp.Diagnostics.AddAttributeError(path.Root("precision"), "Precision and scale require decimal",
			"precision and scale can only be set when type is decimal.")
	case decimal && data.Scale.ValueInt64() > data.Precision.ValueInt64():
		resp.Diagnostics.AddAttributeError(path.Root("scale"), "Scale exceeds precision",
			"scale cannot be greater than precision.")
	}
}

func (r *connectionSchemaFieldResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var config, state, plan connectionSchemaFieldResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planFieldType(ctx, config, state, &plan)
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

// planFieldType plans the type attributes the configuration leaves unset: they
// keep their state while the type is unchanged, and are unknown once it
// changes, except that precision and scale are null for any type but decimal.
func planFieldType(ctx context.Context, config, state connectionSchemaFieldResourceModel, plan *connectionSchemaFieldResourceModel) {
	changed := fieldTypeChanged(ctx, config, state)
	notDecimal := !config.Type.IsNull() && !config.Type.IsUnknown() && config.Type.ValueString() != "decimal"

	if config.Type.IsNull() {
		plan.Type = state.Type
		if changed {
			plan.Type = types.StringUnknown()
		}
	}
	if config.TypeSpec.IsNull() {
		plan.TypeSpec = state.TypeSpec
		if changed {
			plan.TypeSpec = newTypeSpecUnknown()
		}
	}
	for _, attr := range []struct {
		config, state types.Int64
		plan          *types.Int64
	}{
		{config.Precision, state.Precision, &plan.Precision},
		{config.Scale, state.Scale, &plan.Scale},
	} {
		if !attr.config.IsNull() {
			continue
		}
		switch {
		case !changed:
			*attr.plan = attr.state
		case notDecimal:
			*attr.plan = types.Int64Null()
		default:
			*attr.plan = types.Int64Unknown()
		}
	}
}

// fieldTypeChanged reports whether the configuration sets the field's type to
// something other than its current value.
func fieldTypeChanged(ctx context.Context, config, state connectionSchemaFieldResourceModel) bool {
	want := config.TypeSpec
	switch {
	case !want.IsNull():
	case config.Type.IsNull():
		return false
	case state.TypeSpec.IsNull():
		// Without a stored definition, only the type attributes can be compared.
		return !config.Type.Equal(state.Type) ||
			(!config.Precision.IsNull() && !config.Precision.Equal(state.Precision)) ||
			(!config.Scale.IsNull() && !config.Scale.Equal(state.Scale))
	case config.Type.IsUnknown() || config.Precision.IsUnknown() || config.Scale.IsUnknown():
		return true
	default:
		// State reports the basic type of a detailed definition, such as array
		// for ["array", "string"], so compare the definition the configured
		// type stands for instead.
		_, spec, err := fieldTypeSpec(config.Type.ValueString(), config.Precision.ValueInt64(), config.Scale.ValueInt64())
		if err != nil {
			return true
		}
		buf, err := json.Marshal(spec)
		if err != nil {
			return true
		}
		want = newTypeSpecValue(string(buf))
	}
	if want.IsUnknown() || state.TypeSpec.IsNull() || state.TypeSpec.IsUnknown() {
		return true
	}
	equal, diags := want.StringSemanticEquals(ctx, state.TypeSpec)
	return diags.HasError() || !equal
}

// fieldTypeRequest returns the API type and definition for the configured type
// attributes, or an empty type when none are set.
func fieldTypeRequest(m connectionSchemaFieldResourceModel) (string, *polytomic.TypesDefinition, error) {
	var basic string
	var spec any
	switch {
	case !m.TypeSpec.IsNull() && !m.TypeSpec.IsUnknown():
		if err := json.Unmarshal([]byte(m.TypeSpec.ValueString()), &spec); err != nil {
			return "", nil, fmt.Errorf("type_spec: %w", err)
		}
		var err error
		if basic, err = specBasicType(spec); err != nil {
			return "", nil, err
		}
	case !m.Type.IsNull() && !m.Type.IsUnknown():
		var err error
		if basic, spec, err = fieldTypeSpec(m.Type.ValueString(), m.Precision.ValueInt64(), m.Scale.ValueInt64()); err != nil {
			return "", nil, err
		}
	default:
		return "", nil, nil
	}
	def, err := newTypesDefinition(spec)
	return basic, def, err
}

func (r *connectionSchemaFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionSchemaFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	basic, def, err := fieldTypeRequest(data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid field type", err.Error())
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
		if knownStringPointer(data.Label) == nil || basic == "" {
			resp.Diagnostics.AddError(
				"Missing label or type",
				fmt.Sprintf("Schema %s has no field %s, so adding it requires a label and either type or type_spec.", schemaID, fieldID),
			)
			return
		}
		err = client.Schemas.UpsertField(ctx, connectionID, schemaID, &polytomic.UpsertSchemaFieldRequest{
			Fields: []*polytomic.UserFieldRequest{{
				FieldID:    fieldID,
				Label:      data.Label.ValueString(),
				Type:       basic,
				Definition: def,
				Path:       knownStringPointer(data.Path),
			}},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error adding field", err.Error())
			return
		}
		// Adding a field returns no body, so read the merged field back. The
		// field exists once it is added, so record it even when that read
		// fails: otherwise the next apply finds it user-defined and refuses
		// to add it.
		if field, err = readSchemaField(ctx, client, connectionID, schemaID, fieldID); err != nil {
			resp.Diagnostics.AddWarning(
				"Error reading added field",
				fmt.Sprintf("Field %s was added to schema %s, but reading it back failed: %s. "+
					"Attributes the API fills in stay unset until the next refresh.", fieldID, schemaID, err),
			)
		}
	case pointer.GetBool(existing.UserManaged):
		resp.Diagnostics.AddError(
			"Field already has a user-defined definition",
			fmt.Sprintf("Field %s in schema %s is already user-defined or overridden. Import it with the ID %s instead.",
				fieldID, schemaID, SchemaFieldResourceID(resourceOrganization(ctx, client, data.Organization, connectionID), connectionID, schemaID, fieldID)),
		)
		return
	default:
		patch := &polytomic.PatchSchemaFieldRequest{
			Label:      knownStringPointer(data.Label),
			Path:       knownStringPointer(data.Path),
			Definition: def,
		}
		if basic != "" {
			patch.Type = pointer.To(basic)
		}
		if patch.Label == nil && patch.Type == nil && patch.Path == nil {
			resp.Diagnostics.AddError(
				"Nothing to override",
				fmt.Sprintf("Set at least one of label, type, type_spec, or path to override field %s.", fieldID),
			)
			return
		}
		field, err = patchSchemaField(ctx, client, connectionID, schemaID, fieldID, patch)
		if err != nil {
			resp.Diagnostics.AddError("Error overriding field", err.Error())
			return
		}
	}

	if field == nil {
		nullUnknownAttributes(&data)
	} else if err := applySchemaField(&data, field); err != nil {
		resp.Diagnostics.AddError("Error reading field", err.Error())
		return
	}

	data.Organization = types.StringValue(resourceOrganization(ctx, client, data.Organization, connectionID))
	data.ID = types.StringValue(SchemaFieldResourceID(data.Organization.ValueString(), connectionID, schemaID, fieldID))

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
	if err := applySchemaField(&data, field); err != nil {
		resp.Diagnostics.AddError("Error reading field", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchemaFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config, plan, state connectionSchemaFieldResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patch := &polytomic.PatchSchemaFieldRequest{}
	if !plan.Label.Equal(state.Label) {
		patch.Label = knownStringPointer(plan.Label)
	}
	if !plan.Path.Equal(state.Path) {
		patch.Path = knownStringPointer(plan.Path)
	}
	if fieldTypeChanged(ctx, config, state) {
		basic, def, err := fieldTypeRequest(config)
		if err != nil {
			resp.Diagnostics.AddError("Invalid field type", err.Error())
			return
		}
		if basic != "" {
			patch.Type = pointer.To(basic)
		}
		patch.Definition = def
	}

	if patch.Label != nil || patch.Type != nil || patch.Path != nil || patch.Definition != nil {
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
		if err := applySchemaField(&plan, field); err != nil {
			resp.Diagnostics.AddError("Error reading field", err.Error())
			return
		}
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

	connectionID := data.ConnectionID.ValueString()
	schemaID := data.SchemaID.ValueString()
	fieldID := data.FieldID.ValueString()

	err = client.Schemas.DeleteField(ctx, connectionID, schemaID, fieldID)
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Error deleting field", err.Error())
		return
	}

	// The API refreshes the schema in the background after deleting a field,
	// and reports the field as user-defined until then, so adding it again,
	// as replacing this resource does, would fail.
	if err := waitForFieldRemoval(ctx, client, connectionID, schemaID, fieldID); err != nil {
		resp.Diagnostics.AddWarning(
			"Deleted field still reported",
			fmt.Sprintf("Field %s was deleted from schema %s, but the schema did not reflect it within %s: %s. "+
				"Adding the field again fails until the connection's schemas are refreshed.",
				fieldID, schemaID, fieldRemovalTimeout, err),
		)
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

// Schema IDs may contain slashes, so resource IDs escape the field ID to keep
// their last segment unambiguous.
var (
	fieldIDEscaper   = strings.NewReplacer("%", "%25", "/", "%2F")
	fieldIDUnescaper = strings.NewReplacer("%2F", "/", "%2f", "/", "%25", "%")
)

// SchemaFieldResourceID returns the ID of a polytomic_connection_schema_field
// resource: organization/connection_id/schema_id/field_id, with "%" and "/" in
// the field ID escaped as %25 and %2F.
func SchemaFieldResourceID(org, connectionID, schemaID, fieldID string) string {
	return strings.Join([]string{org, connectionID, schemaID, fieldIDEscaper.Replace(fieldID)}, "/")
}

// parseSchemaFieldID splits an identifier built by SchemaFieldResourceID.
func parseSchemaFieldID(id string) (org, connectionID, schemaID, fieldID string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) >= 4 {
		org, connectionID = parts[0], parts[1]
		schemaID = strings.Join(parts[2:len(parts)-1], "/")
		fieldID = fieldIDUnescaper.Replace(parts[len(parts)-1])
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

func readSchemaField(ctx context.Context, client *ptclient.Client, connectionID, schemaID, fieldID string) (*polytomic.SchemaField, error) {
	schemaData, err := fetchSchema(ctx, client, connectionID, schemaID)
	if err != nil {
		return nil, err
	}
	field := findSchemaField(schemaData, fieldID)
	if field == nil {
		return nil, fmt.Errorf("schema %s has no field %s", schemaID, fieldID)
	}
	return field, nil
}

// Variables so tests can shorten them.
var (
	fieldRemovalTimeout  = 5 * time.Minute
	fieldRemovalInterval = 2 * time.Second
)

// waitForFieldRemoval polls a schema until it no longer reports a deleted field
// as user-defined, and returns the last error once fieldRemovalTimeout passes.
func waitForFieldRemoval(ctx context.Context, client *ptclient.Client, connectionID, schemaID, fieldID string) error {
	deadline := time.Now().Add(fieldRemovalTimeout)
	for {
		schemaData, err := fetchSchema(ctx, client, connectionID, schemaID)
		if errors.Is(err, errSchemaNotFound) {
			return nil
		}
		if err == nil {
			field := findSchemaField(schemaData, fieldID)
			if field == nil || !pointer.GetBool(field.UserManaged) {
				return nil
			}
			err = errors.New("the field is still reported as user-defined")
		}
		if time.Now().After(deadline) {
			return err
		}
		tflog.Debug(ctx, "Waiting for deleted field to leave the schema", map[string]any{
			"schema_id": schemaID,
			"field_id":  fieldID,
			"reason":    err.Error(),
		})
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(fieldRemovalInterval):
		}
	}
}

// nullUnknownAttributes nulls the attributes a plan leaves for the API to fill
// in, for recording a field that could not be read back.
func nullUnknownAttributes(data *connectionSchemaFieldResourceModel) {
	for _, s := range []*types.String{&data.Label, &data.Type, &data.Path} {
		if s.IsUnknown() {
			*s = types.StringNull()
		}
	}
	for _, i := range []*types.Int64{&data.Precision, &data.Scale} {
		if i.IsUnknown() {
			*i = types.Int64Null()
		}
	}
	if data.TypeSpec.IsUnknown() {
		data.TypeSpec = newTypeSpecNull()
	}
}

func applySchemaField(data *connectionSchemaFieldResourceModel, f *polytomic.SchemaField) error {
	data.Label = types.StringPointerValue(f.Name)
	data.Path = types.StringPointerValue(f.Path)

	name, precision, scale, _, err := SchemaFieldTypeAttributes(f)
	if err != nil {
		return err
	}
	if name == "" && f.Type != nil {
		name = string(*f.Type)
	}
	data.Type = types.StringNull()
	if name != "" {
		data.Type = types.StringValue(name)
	}
	data.Precision = types.Int64PointerValue(precision)
	data.Scale = types.Int64PointerValue(scale)

	data.TypeSpec = newTypeSpecNull()
	if f.TypeSpec != nil && *f.TypeSpec != nil {
		spec, err := json.Marshal(*f.TypeSpec)
		if err != nil {
			return fmt.Errorf("encoding type_spec for field %s: %w", pointer.GetString(f.ID), err)
		}
		data.TypeSpec = newTypeSpecValue(string(spec))
	}
	return nil
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
