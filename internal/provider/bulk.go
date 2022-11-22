package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type bulkSourceDatasourceData struct {
	ConnectionID types.String `tfsdk:"connection_id"`
	Schemas      types.List   `tfsdk:"schemas"`
}

type bulkDestinationDatasourceData struct {
	ConnectionID          types.String `tfsdk:"connection_id"`
	RequiredConfiguration types.List   `tfsdk:"required_configuration"`
	Modes                 types.List   `tfsdk:"modes"`
}
