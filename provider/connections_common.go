package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/polytomic/terraform-provider-polytomic/provider/internal/connections"
)

var (
	// ConnectionsMap is a map of all the connections that can be imported as
	// resources.
	ConnectionsMap = connectionsMap()

	// ConnectionDatasourcesMap is a map of all the connections that can be
	// imported as data sources.
	ConnectionDatasourcesMap = datasourcesMap()
)

// connectionsMap combines the generated importable connections
// with any additional connections that are not generated.
func connectionsMap() map[string]resource.Resource {
	conns := make(map[string]resource.Resource)
	for k, v := range connections.ImportableResources {
		conns[k] = v
	}

	return conns
}

func datasourcesMap() map[string]datasource.DataSource {
	sources := make(map[string]datasource.DataSource)
	for k, v := range connections.ImportableDatasources {
		sources[k] = v
	}
	return sources
}
