---
page_title: "polytomic_bigquery_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Google BigQuery Connection
---

# polytomic_bigquery_connection (Resource)

Google BigQuery Connection

For detailed configuration guidance, see the [Google BigQuery connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/bigquery).

## Example Usage

### Using WIF

```terraform
resource "polytomic_bigquery_connection" "bigquery" {
  name = "BigQuery (WIF)"

  configuration = {
    auth_method       = "workload_identity_federation"
    wif_project_id    = "my-gcp-project"
    credential_config = file(var.bq_credential_config_path)
  }
}
```

### Using Service Account

```terraform
resource "polytomic_bigquery_connection" "bigquery" {
  name = "BigQuery (Service Account)"

  configuration = {
    auth_method     = "service_account_key"
    project_id      = "my-gcp-project"
    service_account = file(var.bq_service_account_json_key_path)
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Google BigQuery Connection identifier.
- `force_destroy` (Boolean, Optional) Indicates whether dependent models, syncs, and bulk syncs should be
cascade-deleted when this connection is destroyed.

    This only deletes other resources when the connection is destroyed, not when
setting this parameter to `true`. Once this parameter is set to `true`, there
must be a successful `terraform apply` run before a destroy is required to
update this value in the resource state. Without a successful `terraform apply`
after this parameter is set, this flag will have no effect. If setting this
field in the same operation that would require replacing the connection or
destroying the connection, this flag will not work. Additionally when importing
a connection, a successful `terraform apply` is required to set this value in
state before it will take effect on a destroy operation.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

- `auth_method` (String, Required) Authentication method Valid values: <code>service_account_key</code> (Service Account Key), <code>workload_identity_federation</code> (Workload Identity Federation). Default: <code>service_account_key</code>.
- `bucket` (String, Optional) Google Cloud Storage bucket
- `client_email` (String, Optional) Service account identity
- `credential_config` (String, Sensitive, Optional) Credential configuration

    Credential configuration JSON file downloaded from Google Cloud
- `location` (String, Optional) Region or multi-region for query operations
- `override_project_id` (String, Optional) Override project ID

    Override service key's project ID for cross-account access
- `project_id` (String, Optional) Service account project ID
- `service_account` (String, Sensitive, Optional) Service account key
- `structured_values_as_json` (Boolean, Optional) Write object and array values as JSON Default: <code>false</code>.
- `use_extract` (Boolean, Optional) Use Extract for bulk sync from BigQuery
- `wif_project_id` (String, Optional) Google Cloud project ID

