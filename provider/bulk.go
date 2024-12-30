package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type bulkSourceDatasourceData struct {
	ConnectionID types.String `tfsdk:"connection_id"`
	Organization types.String `tfsdk:"organization"`
	Schemas      types.List   `tfsdk:"schemas"`
}

type bulkDestinationDatasourceData struct {
	ConnectionID          types.String `tfsdk:"connection_id"`
	Organization          types.String `tfsdk:"organization"`
	RequiredConfiguration types.Set    `tfsdk:"required_configuration"`
	Modes                 types.Set    `tfsdk:"modes"`
}
