// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &DialpadConnectionDataSource{}

// ExampleDataSource defines the data source implementation.
type DialpadConnectionDataSource struct {
	client *polytomic.Client
}

func (d *DialpadConnectionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dialpad_connection"
}

func (d *DialpadConnectionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Connections: Dialpad Connection",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
			},
			"configuration": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{},
				Optional:   true,
			},
		},
	}
}

func (d *DialpadConnectionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DialpadConnectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
