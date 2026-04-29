---
page_title: "polytomic_gcs_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Google Cloud Storage Connection
---

# polytomic_gcs_connection (Resource)

Google Cloud Storage Connection

For detailed configuration guidance, see the [Google Cloud Storage connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/gcs).

## Example Usage

```terraform
resource "polytomic_gcs_connection" "gcs" {
  name = "example"
  configuration = {
    bucket                   = "my-bucket"
    single_table_file_format = "csv"
    single_table_name        = "collection"
  }
}
```

## Schema

### Required

- `name` (String)
- `configuration` (Attributes) See [below for nested schema](#nestedatt--configuration).

### Optional

- `organization` (String) Organization ID.
- `force_destroy` (Boolean) Indicates whether dependent models, syncs, and bulk syncs should be
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

### Read-Only

- `id` (String) Google Cloud Storage Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `bucket` (String)
- `service_account` (String, Sensitive) Service account key

#### Optional

- `csv_has_headers` (Boolean) CSV files have headers

    Whether CSV files have a header row with field names. Default: <code>true</code>.
- `directory_glob_pattern` (String) Tables glob path
- `is_directory_snapshot` (Boolean) Multi-directory multi-table Default: <code>false</code>.
- `is_single_table` (Boolean) Files are time-based snapshots

    Treat the files as a single table. Default: <code>false</code>.
- `single_table_file_format` (String) File format Valid values: <code>csv</code> (CSV), <code>json</code> (JSON), <code>parquet</code> (Parquet). Default: <code>csv</code>.
- `single_table_file_formats` (Set of String) File formats

    File formats that may be present across different tables Default: <code>[[csv]]</code>.
- `single_table_name` (String) Collection name
- `skip_lines` (Number) Skip first lines

    Skip first N lines of each CSV file. Default: <code>0</code>.

#### Read-Only

- `client_email` (String) Service account identity
- `project_id` (String) Service account project ID


