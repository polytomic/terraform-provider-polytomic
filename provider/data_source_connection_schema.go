package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	ptcore "github.com/polytomic/polytomic-go/core"
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
	ID           types.String `tfsdk:"id"`
}

type schemaFieldModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	IsPrimaryKey types.Bool   `tfsdk:"is_primary_key"`
}

func (d *connectionSchemaDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schema"
}

func (d *connectionSchemaDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Connection Schema Data Source\n\n" +
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
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Field ID",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Field name",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Field type",
							Computed:            true,
						},
						"is_primary_key": schema.BoolAttribute{
							MarkdownDescription: "Whether this field is marked as a primary key",
							Computed:            true,
						},
					},
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

	client, err := d.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	schemaResp, err := client.Schemas.Get(ctx, data.ConnectionID.ValueString(), data.SchemaID.ValueString())
	if err != nil {
		pErr := &ptcore.APIError{}
		if ok := errors.As(err, &pErr); ok {
			if pErr.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddError(
					"Schema not found",
					fmt.Sprintf("Schema %s not found in connection %s", data.SchemaID.ValueString(), data.ConnectionID.ValueString()),
				)
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

	schemaData := schemaResp.Data

	// Set basic attributes
	if schemaData.Name != nil {
		data.Name = types.StringValue(*schemaData.Name)
	} else {
		data.Name = types.StringNull()
	}

	// Convert fields to Terraform model
	fields := []schemaFieldModel{}
	if schemaData.Fields != nil {
		for _, field := range schemaData.Fields {
			fieldModel := schemaFieldModel{}

			if field.Id != nil {
				fieldModel.ID = types.StringValue(*field.Id)
			} else {
				fieldModel.ID = types.StringNull()
			}

			if field.Name != nil {
				fieldModel.Name = types.StringValue(*field.Name)
			} else {
				fieldModel.Name = types.StringNull()
			}

			if field.Type != nil {
				fieldModel.Type = types.StringValue(string(*field.Type))
			} else {
				fieldModel.Type = types.StringNull()
			}

			// Note: The IsPrimaryKey flag is set in the schema field
			// This reflects the current primary key configuration (either auto-detected or overridden)
			fieldModel.IsPrimaryKey = types.BoolValue(false) // Default to false if not present

			fields = append(fields, fieldModel)
		}
	}

	// Convert to types.Set
	fieldsSet, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":             types.StringType,
			"name":           types.StringType,
			"type":           types.StringType,
			"is_primary_key": types.BoolType,
		},
	}, fields)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Fields = fieldsSet

	// Set organization if not provided
	if data.Organization.IsNull() {
		// Try to get organization from connection
		connResp, err := client.Connections.Get(ctx, data.ConnectionID.ValueString())
		if err == nil && connResp.Data != nil && connResp.Data.OrganizationId != nil {
			data.Organization = types.StringValue(*connResp.Data.OrganizationId)
		}
	}

	// Set ID in composite format
	orgID := data.Organization.ValueString()
	if orgID == "" {
		orgID = "default"
	}
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", orgID, data.ConnectionID.ValueString(), data.SchemaID.ValueString()))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
