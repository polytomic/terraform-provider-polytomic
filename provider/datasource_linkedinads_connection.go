package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mitchellh/mapstructure"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &LinkedinAdConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type LinkedinAdConnectionDataSource struct {
	client *polytomic.Client
}

func (d *LinkedinAdConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_linkedinads_connection"
}

func (d *LinkedinAdConnectionDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: LinkedinAd Connection",
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Optional:            true,
			},
			"id": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"organization": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Optional:            true,
			},
			"configuration": {
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"accounts": {
						MarkdownDescription: "",
						Type: types.SetType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"value": types.StringType,
									"label": types.StringType,
								},
							},
						},
						Required:  false,
						Optional:  true,
						Sensitive: false,
					},
				}),
				Optional: true,
			},
		},
	}, nil
}

func (d *LinkedinAdConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *LinkedinAdConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data connectionData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the connection
	connection, err := d.client.Connections().Get(ctx, uuid.MustParse(data.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection", err.Error())
		return
	}

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue(connection.ID)
	data.Name = types.StringValue(connection.Name)
	data.Organization = types.StringValue(connection.OrganizationId)
	var conf polytomic.LinkedinAdsConnectionConfiguration
	err = mapstructure.Decode(connection.Configuration, &conf)
	if err != nil {
		resp.Diagnostics.AddError("Error decoding connection", err.Error())
		return
	}

	var accounts []attr.Value
	for _, acc := range conf.Accounts {
		accounts = append(accounts, types.ObjectValueMust(
			map[string]attr.Type{
				"value": types.StringType,
				"label": types.StringType,
			},
			map[string]attr.Value{
				"value": types.StringValue(acc.Value),
				"label": types.StringValue(*acc.Label),
			},
		))
	}

	var diags diag.Diagnostics
	data.Configuration, diags = types.ObjectValue(
		data.Configuration.AttributeTypes(ctx),
		map[string]attr.Value{
			"accounts": types.SetValueMust(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"value": types.StringType,
						"label": types.StringType,
					},
				},
				accounts,
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