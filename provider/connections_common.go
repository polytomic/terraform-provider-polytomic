package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	ConnectionNotFoundErr = "not found: no connection found (404)"
)

var (
	// ConnectionsMap is a map of all the connections that can be imported as
	// resources.
	ConnectionsMap = connectionsMap()

	// ConnectionDatasourcesMap is a map of all the connections that can be
	// imported as data sources.
	ConnectionDatasourcesMap = datasourcesMap()
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
	for k, v := range connectionImportableResources {
		conns[k] = v
	}
	conns["api"] = &APIConnectionResource{}

	return conns
}

func datasourcesMap() map[string]datasource.DataSource {
	sources := make(map[string]datasource.DataSource)
	for k, v := range connectionImportableDatasources {
		sources[k] = v
	}
	return sources
}
