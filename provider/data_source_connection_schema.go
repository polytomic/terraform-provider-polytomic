package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &connectionSchemaDataSource{}

func NewConnectionSchemaDataSource() datasource.DataSource {
	return &connectionSchemaDataSource{}
}

type connectionSchemaDataSource struct {
	provider *providerclient.Provider
}

type connectionSchemaDataSourceModel struct {
	Organization types.String `tfsdk:"organization"`
	ConnectionID types.String `tfsdk:"connection_id"`
	SchemaID     types.String `tfsdk:"schema_id"`
	Name         types.String `tfsdk:"name"`
	Fields       types.Set    `tfsdk:"fields"`
	FieldsByID   types.Map    `tfsdk:"fields_by_id"`
	ID           types.String `tfsdk:"id"`
}

func (d *connectionSchemaDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schema"
}

func (d *connectionSchemaDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Schemas: Connection Schema Data Source\n\n" +
			"Retrieves information about a connection schema, including its fields and primary key configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier in the format: organization/connection_id/schema_id",
				Computed:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Optional:            true,
				Computed:            true,
			},
			"connection_id": schema.StringAttribute{
				MarkdownDescription: "Connection ID",
				Required:            true,
			},
			"schema_id": schema.StringAttribute{
				MarkdownDescription: "Schema ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Schema name",
				Computed:            true,
			},
			"fields": schema.SetNestedAttribute{
				MarkdownDescription: "Schema fields",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: schemaFieldDataSourceAttributes(),
				},
			},
			"fields_by_id": schema.MapNestedAttribute{
				MarkdownDescription: "Schema fields, keyed by field ID",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: schemaFieldDataSourceAttributes(),
				},
			},
		},
	}
}

func (d *connectionSchemaDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *connectionSchemaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data connectionSchemaDataSourceModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.provider.Client(ctx, data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	schemaData, err := fetchSchema(ctx, client, data.ConnectionID.ValueString(), data.SchemaID.ValueString())
	if errors.Is(err, errSchemaNotFound) {
		resp.Diagnostics.AddError(
			"Schema not found",
			fmt.Sprintf("Schema %s not found in connection %s", data.SchemaID.ValueString(), data.ConnectionID.ValueString()),
		)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading schema", err.Error())
		return
	}

	data.Name = types.StringPointerValue(schemaData.Name)

	fields, err := newSchemaFieldModels(schemaData.Fields)
	if err != nil {
		resp.Diagnostics.AddError("Error reading schema", err.Error())
		return
	}
	fieldsByID := make(map[string]schemaFieldModel, len(fields))
	for _, f := range fields {
		fieldsByID[f.ID.ValueString()] = f
	}

	data.Fields, diags = types.SetValueFrom(ctx, schemaFieldObjectType, fields)
	resp.Diagnostics.Append(diags...)
	data.FieldsByID, diags = types.MapValueFrom(ctx, schemaFieldObjectType, fieldsByID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Organization.IsNull() {
		if org := connectionOrganization(ctx, client, data.ConnectionID.ValueString()); org != "" {
			data.Organization = types.StringValue(org)
		}
	}

	// Set ID in composite format
	orgID := data.Organization.ValueString()
	if orgID == "" {
		orgID = providerclient.DefaultOrganization
	}
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", orgID, data.ConnectionID.ValueString(), data.SchemaID.ValueString()))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
