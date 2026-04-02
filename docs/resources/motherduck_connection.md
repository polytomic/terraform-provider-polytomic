---
page_title: "polytomic_motherduck_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  MotherDuck Connection
---

# polytomic_motherduck_connection (Resource)

MotherDuck Connection

For detailed configuration guidance, see the [MotherDuck connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/motherduck).

## Example Usage

```terraform
resource "polytomic_motherduck_connection" "motherduck" {
  name = "example"
  configuration = {
    aws_access_key_id     = "AKIAIOSFODNN7EXAMPLE"
    aws_secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    database              = "my_db"
    s3_bucket_name        = "s3://polytomic/dataset"
    s3_bucket_region      = "us-east-1"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) MotherDuck Connection identifier.
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

- `access_token` (String, Sensitive, Required) Access Token
- `aws_access_key_id` (String, Optional) AWS Access Key ID (destinations only)

    Access Key ID with read/write access to a bucket.
- `aws_secret_access_key` (String, Sensitive, Optional) AWS Secret Access Key (destinations only)
- `aws_user` (String, Optional) User ARN
- `database` (String, Required)
- `s3_bucket_name` (String, Optional) S3 Bucket Name (destinations only)

    Bucket name (folder optional); ex: s3://polytomic/dataset
- `s3_bucket_region` (String, Optional) S3 Bucket Region (destinations only)

