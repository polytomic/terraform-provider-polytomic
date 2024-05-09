package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &identityDatasource{}

type identityDatasource struct {
	client *polytomic.Client
}

type identityDatasourceData struct {
	ID               types.String `tfsdk:"id"`
	Email            types.String `tfsdk:"email"`
	Role             types.String `tfsdk:"role"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	OrganizationName types.String `tfsdk:"organization_name"`
	IsUser           types.Bool   `tfsdk:"is_user"`
	IsOrganization   types.Bool   `tfsdk:"is_organization"`
	IsPartner        types.Bool   `tfsdk:"is_partner"`
	IsSystem         types.Bool   `tfsdk:"is_system"`
}

func (id *identityDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caller_identity"
}

func (id *identityDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Identity: Caller Identity",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"organization_name": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"is_user": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"is_organization": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"is_partner": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"is_system": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
		},
	}
}

func (id *identityDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	id.client = client
}

func (id *identityDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data identityDatasourceData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the schemas
	identity, err := id.client.Identity().Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error getting identity", err.Error())
		return
	}

	data.ID = types.StringValue(identity.ID.String())
	data.Email = types.StringValue(identity.Email)
	data.Role = types.StringValue(identity.Role)
	data.OrganizationID = types.StringValue(identity.OrganizationID.String())
	data.OrganizationName = types.StringValue(identity.Organization)
	data.IsUser = types.BoolValue(identity.IsUser)
	data.IsOrganization = types.BoolValue(identity.IsOrganization)
	data.IsPartner = types.BoolValue(identity.IsPartner)
	data.IsSystem = types.BoolValue(identity.IsSystem)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
