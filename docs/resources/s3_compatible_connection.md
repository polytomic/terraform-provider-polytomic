---
page_title: "polytomic_s3_compatible_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  S3 Compatible Connection
---

# polytomic_s3_compatible_connection (Resource)

S3 Compatible Connection

For detailed configuration guidance, see the [S3 Compatible connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/s3_compatible).

## Example Usage

```terraform
resource "polytomic_s3_compatible_connection" "s3_compatible" {
  name = "example"
  configuration = {
    aws_access_key_id        = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key    = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    endpoint                 = "https://s3.example.com"
    s3_bucket_name           = "s3://polytomic/dataset"
    s3_bucket_region         = "us-east-1"
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

- `id` (String) S3 Compatible Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `aws_access_key_id` (String) Access key ID

    Access key ID with read/write access to the bucket.
- `aws_secret_access_key` (String, Sensitive) Secret access key
- `endpoint` (String) URL of the S3-compatible API; ex: https://s3.example.com
- `s3_bucket_name` (String) S3 bucket name

    Bucket name (folder optional); ex: s3://polytomic/dataset

#### Optional

- `csv_has_headers` (Boolean) CSV files have headers

    Whether CSV files have a header row with field names. Default: <code>true</code>.
- `directory_glob_pattern` (String) Tables glob path
- `is_directory_snapshot` (Boolean) Multi-directory multi-table Default: <code>false</code>.
- `is_single_table` (Boolean) Files are time-based snapshots

    Treat the files as a single table. Default: <code>false</code>.
- `s3_bucket_region` (String) S3 bucket region

    Region to sign requests for. Most S3-compatible stores ignore this value. Default: <code>us-east-1</code>.
- `single_table_file_format` (String) File format Valid values: <code>csv</code> (CSV), <code>json</code> (JSON), <code>parquet</code> (Parquet). Default: <code>csv</code>.
- `single_table_file_formats` (Set of String) File formats

    File formats that may be present across different tables Default: <code>[[csv]]</code>.
- `single_table_name` (String) Collection name
- `skip_lines` (Number) Skip first lines

    Skip first N lines of each CSV file. Default: <code>0</code>.


