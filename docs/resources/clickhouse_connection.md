---
page_title: "polytomic_clickhouse_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  ClickHouse Connection
---

# polytomic_clickhouse_connection (Resource)

ClickHouse Connection

For detailed configuration guidance, see the [ClickHouse connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/clickhouse).

## Example Usage

```terraform
resource "polytomic_clickhouse_connection" "clickhouse" {
  name = "example"
  configuration = {
    auth_mode             = "access_key_and_secret"
    aws_access_key_id     = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    azure_access_key      = "abcdefghijklmnopqrstuvwxyz0123456789/+ABCDEabcdefghijklmnopqrstuvwxyz0123456789/+ABCDE=="
    azure_account_name    = "account"
    cloud_provider        = "aws"
    container_name        = "container"
    database              = "default"
    hostname              = "clickhouse.example.com"
    s3_bucket_name        = "my-bucket"
    s3_bucket_region      = "us-east-1"
    ssh_host              = "bastion.example.com"
    username              = "default"
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

- `id` (String) ClickHouse Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `hostname` (String)
- `port` (Number) Default: <code>9440</code>.
- `username` (String)

#### Optional

- `auth_mode` (String) AWS Authentication Method

    How to authenticate with AWS for the staging bucket Valid values: <code>access_key_and_secret</code> (Access Key and Secret), <code>iam_role</code> (IAM role). Default: <code>access_key_and_secret</code>.
- `aws_access_key_id` (String) AWS Access Key ID (destinations only)
- `aws_secret_access_key` (String, Sensitive) AWS Secret Access Key (destinations only)
- `azure_access_key` (String, Sensitive) Storage Account Access Key (destinations only)
- `azure_account_name` (String) Storage Account Name (destinations only)
- `cloud_provider` (String) Cloud Provider (destination support only) Valid values: <code>aws</code> (AWS), <code>azure</code> (Azure).
- `container_name` (String) Storage Container Name (destinations only)

    Container used for staging data load files (may be "container" or "container/prefix")
- `database` (String)
- `iam_role_arn` (String) IAM Role ARN
- `password` (String, Sensitive)
- `s3_bucket_name` (String) S3 Bucket Name (destinations only)

    Name of bucket used for staging data load files
- `s3_bucket_region` (String) S3 Bucket Region (destinations only)
- `skip_verify` (Boolean) Skip certificate verification Default: <code>true</code>.
- `ssh` (Boolean) Connect over SSH tunnel
- `ssh_host` (String) SSH host
- `ssh_port` (Number) SSH port Default: <code>22</code>.
- `ssh_private_key` (String, Sensitive) Private key
- `ssh_user` (String) SSH user Default: <code>root</code>.
- `ssl` (Boolean) Use SSL Default: <code>true</code>.

#### Read-Only

- `aws_user` (String) User ARN (destinations only)
- `external_id` (String) External ID

    External ID for the IAM role


