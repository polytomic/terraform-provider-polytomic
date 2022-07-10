package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type connectionResourceData struct {
	Workspace     types.String `tfsdk:"workspace"`
	Name          types.String `tfsdk:"name"`
	Id            types.String `tfsdk:"id"`
	Configuration types.Object `tfsdk:"configuration"`
}
