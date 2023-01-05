package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	ConnectionNotFoundErr = "no connection found (404)"
)

var (
	ConnectionsMap = map[string]resource.Resource{
		"affinity":   &AffinityConnectionResource{},
		"airtable":   &AirtableConnectionResource{},
		"api":        &APIConnectionResource{},
		"amplitude":  &AmplitudeConnectionResource{},
		"athena":     &AthenaConnectionResource{},
		"azureblob":  &AzureblobConnectionResource{},
		"bigquery":   &BigqueryConnectionResource{},
		"gcs":        &GcsConnectionResource{},
		"marketo":    &MarketoConnectionResource{},
		"mongodb":    &MongodbConnectionResource{},
		"postgresql": &PostgresqlConnectionResource{},
		"s3":         &S3ConnectionResource{},
		"snowflake":  &SnowflakeConnectionResource{},
		"sqlserver":  &SqlserverConnectionResource{},
	}
)

type connectionData struct {
	Organization  types.String `tfsdk:"organization"`
	Name          types.String `tfsdk:"name"`
	Id            types.String `tfsdk:"id"`
	Configuration types.Object `tfsdk:"configuration"`
}
