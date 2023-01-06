// Code generated by Polytomic. DO NOT EDIT.
// edit connections.yaml and re-run go generate

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	connectionResources = []func() resource.Resource{
		func() resource.Resource { return &PostgresqlConnectionResource{} },
		func() resource.Resource { return &BigqueryConnectionResource{} },
		func() resource.Resource { return &GcsConnectionResource{} },
		func() resource.Resource { return &AzureblobConnectionResource{} },
		func() resource.Resource { return &S3ConnectionResource{} },
		func() resource.Resource { return &SqlserverConnectionResource{} },
		func() resource.Resource { return &AthenaConnectionResource{} },
		func() resource.Resource { return &SnowflakeConnectionResource{} },
		func() resource.Resource { return &AffinityConnectionResource{} },
		func() resource.Resource { return &AirtableConnectionResource{} },
		func() resource.Resource { return &AmplitudeConnectionResource{} },
		func() resource.Resource { return &MarketoConnectionResource{} },
		func() resource.Resource { return &MongodbConnectionResource{} },
		func() resource.Resource { return &ChargebeeConnectionResource{} },
		func() resource.Resource { return &CloudsqlConnectionResource{} },
		func() resource.Resource { return &CosmosdbConnectionResource{} },
		func() resource.Resource { return &CustomerioConnectionResource{} },
		func() resource.Resource { return &DialpadConnectionResource{} },
		func() resource.Resource { return &FreshdeskConnectionResource{} },
		func() resource.Resource { return &FullstoryConnectionResource{} },
		func() resource.Resource { return &HarmonicConnectionResource{} },
		func() resource.Resource { return &IntercomConnectionResource{} },
		func() resource.Resource { return &KlaviyoConnectionResource{} },
		func() resource.Resource { return &KustomerConnectionResource{} },
		func() resource.Resource { return &LobConnectionResource{} },
		func() resource.Resource { return &MysqlConnectionResource{} },
		func() resource.Resource { return &NetsuiteConnectionResource{} },
		func() resource.Resource { return &PipedriveConnectionResource{} },
		func() resource.Resource { return &RedshiftConnectionResource{} },
		func() resource.Resource { return &SegmentConnectionResource{} },
		func() resource.Resource { return &StripeConnectionResource{} },
	}

	connectionDatasources = []func() datasource.DataSource{
		func() datasource.DataSource { return &PostgresqlConnectionDataSource{} },
		func() datasource.DataSource { return &BigqueryConnectionDataSource{} },
		func() datasource.DataSource { return &GcsConnectionDataSource{} },
		func() datasource.DataSource { return &AzureblobConnectionDataSource{} },
		func() datasource.DataSource { return &S3ConnectionDataSource{} },
		func() datasource.DataSource { return &SqlserverConnectionDataSource{} },
		func() datasource.DataSource { return &AthenaConnectionDataSource{} },
		func() datasource.DataSource { return &SalesforceConnectionDataSource{} },
		func() datasource.DataSource { return &HubspotConnectionDataSource{} },
		func() datasource.DataSource { return &GoogleadsConnectionDataSource{} },
		func() datasource.DataSource { return &GsheetsConnectionDataSource{} },
		func() datasource.DataSource { return &SnowflakeConnectionDataSource{} },
		func() datasource.DataSource { return &AffinityConnectionDataSource{} },
		func() datasource.DataSource { return &AirtableConnectionDataSource{} },
		func() datasource.DataSource { return &AmplitudeConnectionDataSource{} },
		func() datasource.DataSource { return &MarketoConnectionDataSource{} },
		func() datasource.DataSource { return &MongodbConnectionDataSource{} },
		func() datasource.DataSource { return &ChargebeeConnectionDataSource{} },
		func() datasource.DataSource { return &CloudsqlConnectionDataSource{} },
		func() datasource.DataSource { return &CosmosdbConnectionDataSource{} },
		func() datasource.DataSource { return &CustomerioConnectionDataSource{} },
		func() datasource.DataSource { return &DialpadConnectionDataSource{} },
		func() datasource.DataSource { return &FreshdeskConnectionDataSource{} },
		func() datasource.DataSource { return &FullstoryConnectionDataSource{} },
		func() datasource.DataSource { return &HarmonicConnectionDataSource{} },
		func() datasource.DataSource { return &IntercomConnectionDataSource{} },
		func() datasource.DataSource { return &KlaviyoConnectionDataSource{} },
		func() datasource.DataSource { return &KustomerConnectionDataSource{} },
		func() datasource.DataSource { return &LobConnectionDataSource{} },
		func() datasource.DataSource { return &MysqlConnectionDataSource{} },
		func() datasource.DataSource { return &NetsuiteConnectionDataSource{} },
		func() datasource.DataSource { return &PipedriveConnectionDataSource{} },
		func() datasource.DataSource { return &RedshiftConnectionDataSource{} },
		func() datasource.DataSource { return &SegmentConnectionDataSource{} },
		func() datasource.DataSource { return &StripeConnectionDataSource{} },
	}

	connectionImportables = map[string]resource.Resource{
		"postgresql": &PostgresqlConnectionResource{},
		"bigquery":   &BigqueryConnectionResource{},
		"gcs":        &GcsConnectionResource{},
		"azureblob":  &AzureblobConnectionResource{},
		"s3":         &S3ConnectionResource{},
		"sqlserver":  &SqlserverConnectionResource{},
		"athena":     &AthenaConnectionResource{},
		"snowflake":  &SnowflakeConnectionResource{},
		"affinity":   &AffinityConnectionResource{},
		"airtable":   &AirtableConnectionResource{},
		"amplitude":  &AmplitudeConnectionResource{},
		"marketo":    &MarketoConnectionResource{},
		"mongodb":    &MongodbConnectionResource{},
		"chargebee":  &ChargebeeConnectionResource{},
		"cloudsql":   &CloudsqlConnectionResource{},
		"cosmosdb":   &CosmosdbConnectionResource{},
		"customerio": &CustomerioConnectionResource{},
		"dialpad":    &DialpadConnectionResource{},
		"freshdesk":  &FreshdeskConnectionResource{},
		"fullstory":  &FullstoryConnectionResource{},
		"harmonic":   &HarmonicConnectionResource{},
		"intercom":   &IntercomConnectionResource{},
		"klaviyo":    &KlaviyoConnectionResource{},
		"kustomer":   &KustomerConnectionResource{},
		"lob":        &LobConnectionResource{},
		"mysql":      &MysqlConnectionResource{},
		"netsuite":   &NetsuiteConnectionResource{},
		"pipedrive":  &PipedriveConnectionResource{},
		"redshift":   &RedshiftConnectionResource{},
		"segment":    &SegmentConnectionResource{},
		"stripe":     &StripeConnectionResource{},
	}
)
