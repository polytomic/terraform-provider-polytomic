// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/terraform-provider-polytomic/provider/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &CsvConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type CsvConnectionDataSource struct {
	provider *client.Provider
}

func (d *CsvConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := client.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *CsvConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_csv_connection"
}

func (d *CsvConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: CSV URL Connection",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"configuration": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"auth": schema.SingleNestedAttribute{
						MarkdownDescription: "",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"basic": schema.SingleNestedAttribute{
								MarkdownDescription: "",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"password": schema.StringAttribute{
										MarkdownDescription: "",
										Computed:            true,
									},
									"username": schema.StringAttribute{
										MarkdownDescription: "",
										Computed:            true,
									},
								},
							},
							"header": schema.SingleNestedAttribute{
								MarkdownDescription: "",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: "",
										Computed:            true,
									},
									"value": schema.StringAttribute{
										MarkdownDescription: "",
										Computed:            true,
									},
								},
							},
							"oauth": schema.SingleNestedAttribute{
								MarkdownDescription: "",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"auth_style": schema.Int64Attribute{
										MarkdownDescription: "",
										Computed:            true,
									},
									"client_id": schema.StringAttribute{
										MarkdownDescription: "",
										Computed:            true,
									},
									"client_secret": schema.StringAttribute{
										MarkdownDescription: "",
										Computed:            true,
									},
									"extra_form_data": schema.SetNestedAttribute{
										MarkdownDescription: "",
										Computed:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													MarkdownDescription: "",
													Computed:            true,
												},
												"value": schema.StringAttribute{
													MarkdownDescription: "",
													Computed:            true,
												},
											},
										},
									},
									"scopes": schema.SetAttribute{
										MarkdownDescription: "",
										Computed:            true,
										ElementType:         types.StringType,
									},
									"token_endpoint": schema.StringAttribute{
										MarkdownDescription: "",
										Computed:            true,
									},
								},
							},
						},
					},
					"headers": schema.SetNestedAttribute{
						MarkdownDescription: "",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "",
									Computed:            true,
								},
								"value": schema.StringAttribute{
									MarkdownDescription: "",
									Computed:            true,
								},
							},
						},
					},
					"parameters": schema.SetNestedAttribute{
						MarkdownDescription: "",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "",
									Computed:            true,
								},
								"value": schema.StringAttribute{
									MarkdownDescription: "",
									Computed:            true,
								},
							},
						},
					},
					"url": schema.StringAttribute{
						MarkdownDescription: "e.g. http://www.example.com",
						Computed:            true,
					},
				},
				Optional: true,
			},
		},
	}
}

func (d *CsvConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data connectionData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the connection
	client, err := d.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	connection, err := client.Connections.Get(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection", err.Error())
		return
	}

	data.Id = types.StringPointerValue(connection.Data.Id)
	data.Name = types.StringPointerValue(connection.Data.Name)
	data.Organization = types.StringPointerValue(connection.Data.OrganizationId)
	var diags diag.Diagnostics
	data.Configuration, diags = types.ObjectValue(
		data.Configuration.AttributeTypes(ctx),
		map[string]attr.Value{
			"auth": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["auth"], "string").(string),
			),
			"headers": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["headers"], "string").(string),
			),
			"parameters": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["parameters"], "string").(string),
			),
			"url": types.StringValue(
				getValueOrEmpty(connection.Data.Configuration["url"], "string").(string),
			),
		},
	)

	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
