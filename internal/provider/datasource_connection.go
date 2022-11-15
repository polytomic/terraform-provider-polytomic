package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type connectionDatasourceData struct {
	Name          types.String  `tfsdk:"name"`
	ID            types.String  `tfsdk:"id"`
	Organization  types.String  `tfsdk:"organization"`
	Configuration types.MapType `tfsdk:"configuration"`
}
