package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type connectionResourceData struct {
	Organization  types.String `tfsdk:"organization"`
	Name          types.String `tfsdk:"name"`
	Id            types.String `tfsdk:"id"`
	Configuration types.Object `tfsdk:"configuration"`
}
