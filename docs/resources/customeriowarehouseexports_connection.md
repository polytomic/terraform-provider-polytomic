---
page_title: "polytomic_customeriowarehouseexports_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Customer.io Warehouse Exports Connection
---

# polytomic_customeriowarehouseexports_connection (Resource)

Customer.io Warehouse Exports Connection

For detailed configuration guidance, see the [Customer.io Warehouse Exports connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/customeriowarehouseexports).

## Example Usage

```terraform
resource "polytomic_customeriowarehouseexports_connection" "customeriowarehouseexports" {
  name = "example"
  configuration = {
    auth_mode             = "access_key_and_secret"
    aws_access_key_id     = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    s3_bucket_name        = "s3://polytomic/dataset"
    s3_bucket_region      = "us-east-1"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Customer.io Warehouse Exports Connection identifier.
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
- `external_id` (String, Optional) External ID

    External ID for the IAM role
- `iam_role_arn` (String, Optional) IAM Role ARN
- `s3_bucket_name` (String, Required) S3 Bucket Name

    Bucket name (folder optional); ex: s3://polytomic/dataset
- `s3_bucket_region` (String, Required) S3 Bucket Region

