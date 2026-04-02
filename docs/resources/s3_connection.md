---
page_title: "polytomic_s3_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  S3 Connection
---

# polytomic_s3_connection (Resource)

S3 Connection

For detailed configuration guidance, see the [S3 connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/s3).

## Example Usage

```terraform
resource "polytomic_s3_connection" "s3" {
  name = "example"
  configuration = {
    auth_mode                = "access_key_and_secret"
    aws_access_key_id        = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key    = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    event_queue_arn          = "arn:aws:sqs:us-east-1:123456789012:my-queue"
    s3_bucket_name           = "s3://polytomic/dataset"
    s3_bucket_region         = "us-east-1"
    single_table_file_format = "csv"
    single_table_name        = "collection"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) S3 Connection identifier.
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

- `auth_mode` (String, Required) Authentication Method

    How to authenticate with AWS. Defaults to Access Key and Secret Valid values: <code>access_key_and_secret</code> (Access Key and Secret), <code>iam_role</code> (IAM role). Default: <code>access_key_and_secret</code>.
- `aws_access_key_id` (String, Optional) AWS Access Key ID

    Access Key ID with read/write access to a bucket.
- `aws_secret_access_key` (String, Sensitive, Optional) AWS Secret Access Key
- `aws_user` (String, Optional) User ARN
- `csv_has_headers` (Boolean, Optional) CSV files have headers

    Whether CSV files have a header row with field names. Default: <code>true</code>.
- `directory_glob_pattern` (String, Optional) Tables glob path
- `enable_event_notifications` (Boolean, Optional) Enable event notifications

    Enable S3 event notifications for incremental sync
- `event_queue_arn` (String, Optional) Event Queue ARN

    ARN of the SQS queue receiving S3 event notifications
- `external_id` (String, Optional) External ID

    External ID for the IAM role
- `iam_role_arn` (String, Optional) IAM Role ARN
- `is_directory_snapshot` (Boolean, Optional) Multi-directory multi-table Default: <code>false</code>.
- `is_single_table` (Boolean, Optional) Files are time-based snapshots

    Treat the files as a single table. Default: <code>false</code>.
- `s3_bucket_name` (String, Required) S3 Bucket Name

    Bucket name (folder optional); ex: s3://polytomic/dataset
- `s3_bucket_region` (String, Required) S3 Bucket Region
- `single_table_file_format` (String, Optional) File format Valid values: <code>csv</code> (CSV), <code>json</code> (JSON), <code>parquet</code> (Parquet). Default: <code>csv</code>.
- `single_table_file_formats` (Set of String, Optional) File formats

    File formats that may be present across different tables Default: <code>[[csv]]</code>.
- `single_table_name` (String, Optional) Collection name
- `skip_lines` (Number, Optional) Skip first lines

    Skip first N lines of each CSV file. Default: <code>0</code>.

