package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlekSi/pointer"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

var _ datasource.DataSource = &roleDatasource{}

type roleDatasource struct {
	provider *providerclient.Provider
}

type roleDatasourceData struct {
	Organization types.String `tfsdk:"organization"`
	Name         types.String `tfsdk:"name"`
	ID           types.String `tfsdk:"id"`
}

func (d *roleDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *roleDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *roleDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Permissions: Look up a role by name",
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Role name (case-insensitive)",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Role ID",
				Computed:            true,
			},
		},
	}
}

func (d *roleDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data roleDatasourceData

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.provider.Client(data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}

	roles, err := client.Permissions.Roles.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing roles", fmt.Sprintf("Failed to list roles: %s", err))
		return
	}

	for _, role := range roles.Data {
		if strings.EqualFold(pointer.GetString(role.Name), data.Name.ValueString()) {
			data.ID = types.StringPointerValue(role.Id)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Role not found",
		fmt.Sprintf("No role with name %q found in organization %s", data.Name.ValueString(), data.Organization.ValueString()),
	)
}
