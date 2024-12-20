package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	// ConnectionsMap is a map of all the connections that can be imported as
	// resources.
	ConnectionsMap = connectionsMap()

	// ConnectionDatasourcesMap is a map of all the connections that can be
	// imported as data sources.
	ConnectionDatasourcesMap = datasourcesMap()

	forceDestroyMessage = "Indicates whether dependent models, syncs, and bulk syncs should be cascade deleted when this connection is destroy. " +
		"This only deletes other resources when the connection is destroyed, not when setting this parameter to `true`. " +
		"Once this parameter is set to `true`, there must be a successful `terraform apply` run before a destroy is required to update this " +
		"value in the resource state. Without a successful `terraform apply` after this parameter is set, this flag will have no effect. " +
		"If setting this field in the same operation that would require replacing the connection or destroying the connection, this flag will not " +
		"work. Additionally when importing a connection, a successful `terraform apply` is required to set this value in state before it will take effect on a destroy operation."
)

type connectionData struct {
	Organization  types.String `tfsdk:"organization"`
	Name          types.String `tfsdk:"name"`
	Id            types.String `tfsdk:"id"`
	Configuration types.Object `tfsdk:"configuration"`
	ForceDestroy  types.Bool   `tfsdk:"force_destroy"`
}

// connectionsMap combines the generated importable connections
// with any additional connections that are not generated.
func connectionsMap() map[string]resource.Resource {
	conns := make(map[string]resource.Resource)
	for k, v := range connectionImportableResources {
		conns[k] = v
	}

	return conns
}

func datasourcesMap() map[string]datasource.DataSource {
	sources := make(map[string]datasource.DataSource)
	for k, v := range connectionImportableDatasources {
		sources[k] = v
	}
	return sources
}
