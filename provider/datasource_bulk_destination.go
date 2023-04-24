package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &bulkDestinationDatasource{}

// ExampleDataSource defines the data source implementation.
type bulkDestinationDatasource struct {
	client *polytomic.Client
}

func (d *bulkDestinationDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_destination"
}

func (d *bulkDestinationDatasource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: ":meta:subcategory:Bulk Syncs: Bulk Destination",
		Attributes: map[string]tfsdk.Attribute{
			"connection_id": {
				MarkdownDescription: "",
				Type:                types.StringType,
				Required:            true,
			},
			"required_configuration": {
				MarkdownDescription: "",
				Type:                types.SetType{ElemType: types.StringType},
				Computed:            true,
			},
			"modes": {
				MarkdownDescription: "",
				Type: types.SetType{ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":          types.StringType,
						"label":       types.StringType,
						"description": types.StringType,
					},
				}},
				Computed: true,
			},
		},
	}, nil
}

func (d *bulkDestinationDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bulkDestinationDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data bulkDestinationDatasourceData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	dest, err := d.client.Bulk().GetDestination(ctx, data.ConnectionID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection", err.Error())
		return
	}
	var diags diag.Diagnostics
	data.Modes, diags = types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":          types.StringType,
			"label":       types.StringType,
			"description": types.StringType,
		},
	}, dest.Modes)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	raw := dest.Configuration.(map[string]interface{})
	required := make([]string, len(raw))
	i := 0
	for k := range raw {
		required[i] = k
		i++
	}

	data.RequiredConfiguration, diags = types.SetValueFrom(ctx, types.StringType, required)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
