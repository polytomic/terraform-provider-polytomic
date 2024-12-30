package provider

import (
	"context"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &bulkSourceDatasource{}

// ExampleDataSource defines the data source implementation.
type bulkSourceDatasource struct {
	provider *client.Provider
}

func (d *bulkSourceDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *bulkSourceDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_source"
}

func (d *bulkSourceDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Bulk Syncs: Bulk Source",
		Attributes: map[string]schema.Attribute{
			"connection_id": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"schemas": schema.ListAttribute{
				MarkdownDescription: "",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":   types.StringType,
						"name": types.StringType,
						"fields": types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"id":   types.StringType,
									"name": types.StringType,
									"type": types.StringType,
								},
							},
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *bulkSourceDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data bulkSourceDatasourceData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the schemas
	client, err := d.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	source, err := client.BulkSync.GetSource(ctx, data.ConnectionID.ValueString(), &polytomic.BulkSyncGetSourceRequest{})
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection", err.Error())
		return
	}

	schemas := make([]sourceSchema, len(source.Data.Schemas))
	for i, s := range source.Data.Schemas {
		fields := make([]sourceSchemaField, len(s.Fields))
		for j, f := range s.Fields {
			fields[j] = sourceSchemaField{
				ID:   pointer.Get(f.Id),
				Name: pointer.Get(f.Name),
				Type: string(pointer.Get(f.Type)),
			}
		}
		schemas[i] = sourceSchema{
			ID:     pointer.Get(s.Id),
			Name:   pointer.Get(s.Name),
			Fields: fields,
		}
	}
	var diags diag.Diagnostics
	data.Schemas, diags = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":   types.StringType,
			"name": types.StringType,
			"fields": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":   types.StringType,
						"name": types.StringType,
						"type": types.StringType,
					},
				},
			},
		}}, schemas)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type sourceSchema struct {
	ID     string              `json:"id" tfsdk:"id"`
	Name   string              `json:"name" tfsdk:"name"`
	Fields []sourceSchemaField `json:"fields" tfsdk:"fields"`
}

type sourceSchemaField struct {
	ID   string `json:"id" tfsdk:"id"`
	Name string `json:"name" tfsdk:"name"`
	Type string `json:"type" tfsdk:"type"`
}
