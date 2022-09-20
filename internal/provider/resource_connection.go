package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	ConnectionNotFoundErr = "no connection found (404)"
)

type connectionResourceData struct {
	Organization  types.String `tfsdk:"organization"`
	Name          types.String `tfsdk:"name"`
	Id            types.String `tfsdk:"id"`
	Configuration types.Object `tfsdk:"configuration"`
}
