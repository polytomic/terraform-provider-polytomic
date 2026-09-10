package provider

import (
	"context"
	"fmt"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go/v25"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &connectionSchemasDataSource{}

func NewConnectionSchemasDataSource() datasource.DataSource {
	return &connectionSchemasDataSource{}
}

type connectionSchemasDataSource struct {
	provider *providerclient.Provider
}

type connectionSchemasDataSourceModel struct {
	Organization  types.String `tfsdk:"organization"`
	ConnectionID  types.String `tfsdk:"connection_id"`
	IncludeFields types.Bool   `tfsdk:"include_fields"`
	Schemas       types.List   `tfsdk:"schemas"`
	ID            types.String `tfsdk:"id"`
}

type connectionSchemasSchemaModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Fields types.Set    `tfsdk:"fields"`
}

var connectionSchemasSchemaObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":     types.StringType,
		"name":   types.StringType,
		"fields": types.SetType{ElemType: schemaFieldObjectType},
	},
}

func (d *connectionSchemasDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schemas"
}

func (d *connectionSchemasDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Connection Schemas Data Source\n\n" +
			"Lists the schemas available on a connection, optionally with their fields.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier in the format: organization/connection_id",
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
			"include_fields": schema.BoolAttribute{
				MarkdownDescription: "Whether to include each schema's fields. Defaults to `false`; fields can make the response large on connections with many schemas.",
				Optional:            true,
			},
			"schemas": schema.ListNestedAttribute{
				MarkdownDescription: "Schemas available on the connection",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Schema ID",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Schema name",
							Computed:            true,
						},
						"fields": schema.SetNestedAttribute{
							MarkdownDescription: "Schema fields; null unless `include_fields` is `true`",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: schemaFieldDataSourceAttributes(),
							},
						},
					},
				},
			},
		},
	}
}

func (d *connectionSchemasDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *connectionSchemasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data connectionSchemasDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.provider.Client(ctx, data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	includeFields := data.IncludeFields.ValueBool()
	source, err := client.BulkSync.GetSource(ctx, data.ConnectionID.ValueString(), &polytomic.BulkSyncGetSourceRequest{
		IncludeFields: pointer.To(includeFields),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error listing schemas", err.Error())
		return
	}

	schemas := []connectionSchemasSchemaModel{}
	if source.Data != nil {
		for _, s := range source.Data.Schemas {
			if s == nil {
				continue
			}
			m := connectionSchemasSchemaModel{
				ID:     types.StringPointerValue(s.ID),
				Name:   types.StringPointerValue(s.Name),
				Fields: types.SetNull(schemaFieldObjectType),
			}
			if includeFields {
				fields, err := newSchemaFieldModels(s.Fields)
				if err != nil {
					resp.Diagnostics.AddError("Error reading schema fields", err.Error())
					return
				}
				var diags diag.Diagnostics
				m.Fields, diags = types.SetValueFrom(ctx, schemaFieldObjectType, fields)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
			}
			schemas = append(schemas, m)
		}
	}

	var diags diag.Diagnostics
	data.Schemas, diags = types.ListValueFrom(ctx, connectionSchemasSchemaObjectType, schemas)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Organization.IsNull() {
		if org := connectionOrganization(ctx, client, data.ConnectionID.ValueString()); org != "" {
			data.Organization = types.StringValue(org)
		}
	}

	orgID := data.Organization.ValueString()
	if orgID == "" {
		orgID = "default"
	}
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", orgID, data.ConnectionID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
