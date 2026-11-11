package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/polytomic/polytomic-go/v25"
	"github.com/polytomic/terraform-provider-polytomic/internal/providerclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &bulkDestinationDatasource{}

// ExampleDataSource defines the data source implementation.
type bulkDestinationDatasource struct {
	provider *providerclient.Provider
}

func (d *bulkDestinationDatasource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if provider := providerclient.GetProvider(req.ProviderData, resp.Diagnostics); provider != nil {
		d.provider = provider
	}
}

func (d *bulkDestinationDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bulk_destination"
}

func (d *bulkDestinationDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: ":meta:subcategory:Bulk Syncs: Bulk Destination",
		Attributes: map[string]schema.Attribute{
			"source_connection_id": schema.StringAttribute{
				MarkdownDescription: "Source connection used to suggest initial destination configuration values.",
				Optional:            true,
			},
			"initial_configuration": schema.StringAttribute{
				MarkdownDescription: "Suggested initial destination configuration as JSON. Requires source_connection_id; preserve explicit choices and existing configurations.",
				Computed:            true,
			},
			"supports_history_mode": schema.BoolAttribute{
				MarkdownDescription: "Whether the destination supports companion history tables in replicate mode.",
				Computed:            true,
			},
			"connection_id": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"required_configuration": schema.SetAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"modes": schema.SetAttribute{
				MarkdownDescription: "",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":          types.StringType,
						"label":       types.StringType,
						"description": types.StringType,
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *bulkDestinationDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data bulkDestinationDatasourceData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.provider.Client(ctx, data.Organization.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting client", err.Error())
		return
	}
	dest, err := client.BulkSync.GetDestination(ctx, data.ConnectionID.ValueString(), &polytomic.BulkSyncGetDestinationRequest{
		SourceConnectionID: data.SourceConnectionID.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error getting connection", err.Error())
		return
	}
	data.SupportsHistoryMode = types.BoolPointerValue(dest.Data.SupportsHistoryMode)
	data.InitialConfiguration = types.StringNull()
	if dest.Data.InitialConfiguration != nil {
		configuration, err := json.Marshal(dest.Data.InitialConfiguration)
		if err != nil {
			resp.Diagnostics.AddError("Error reading initial configuration", err.Error())
			return
		}
		data.InitialConfiguration = types.StringValue(string(configuration))
	}
	var diags diag.Diagnostics
	modeTypes := map[string]attr.Type{
		"id":          types.StringType,
		"label":       types.StringType,
		"description": types.StringType,
	}
	modes := make([]attr.Value, len(dest.Data.Modes))
	for i, mode := range dest.Data.Modes {
		modes[i] = types.ObjectValueMust(modeTypes, map[string]attr.Value{
			"id":          types.StringPointerValue((*string)(mode.ID)),
			"label":       types.StringPointerValue(mode.Label),
			"description": types.StringPointerValue(mode.Description),
		})
	}
	data.Modes, diags = types.SetValue(types.ObjectType{AttrTypes: modeTypes}, modes)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	raw := dest.Data.Configuration
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
