package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	ConnectionNotFoundErr = "not found: no connection found (404)"
)

var (
	// ConnectionsMap is a map of all the connections that can be imported
	ConnectionsMap = connectionsMap()
)

type connectionData struct {
	Organization  types.String `tfsdk:"organization"`
	Name          types.String `tfsdk:"name"`
	Id            types.String `tfsdk:"id"`
	Configuration types.Object `tfsdk:"configuration"`
}

// connectionsMap combines the generated importable connections
// with any additional connections that are not generated.
func connectionsMap() map[string]resource.Resource {
	conns := make(map[string]resource.Resource)
	for k, v := range connectionImportables {
		conns[k] = v
	}
	conns["api"] = &APIConnectionResource{}

	return conns
}
