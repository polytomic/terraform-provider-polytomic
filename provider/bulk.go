package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type bulkSourceDatasourceData struct {
	ConnectionID types.String `tfsdk:"connection_id"`
	Schemas      types.Set    `tfsdk:"schemas"`
}

type bulkDestinationDatasourceData struct {
	ConnectionID          types.String `tfsdk:"connection_id"`
	RequiredConfiguration types.Set    `tfsdk:"required_configuration"`
	Modes                 types.Set    `tfsdk:"modes"`
}
