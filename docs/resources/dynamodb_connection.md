---
page_title: "polytomic_dynamodb_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  DynamoDB Connection
---

# polytomic_dynamodb_connection (Resource)

DynamoDB Connection

For detailed configuration guidance, see the [DynamoDB connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/dynamodb).

## Example Usage

```terraform
resource "polytomic_dynamodb_connection" "dynamodb" {
  name = "example"
  configuration = {
    access_id         = "AKIAIOSFODNN7EXAMPLE"
    auth_mode         = "access_key_and_secret"
    region            = "us-east-1"
    secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
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

- `id` (String) DynamoDB Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `auth_mode` (String) Authentication Method

    How to authenticate with AWS. Defaults to Access Key and Secret Valid values: <code>access_key_and_secret</code> (Access Key and Secret), <code>iam_role</code> (IAM role). Default: <code>access_key_and_secret</code>.
- `region` (String) AWS region

#### Optional

- `access_id` (String, Sensitive) AWS Access ID
- `change_detection` (Boolean) Use DynamoDB Streams for bulk syncs Default: <code>false</code>.
- `iam_role_arn` (String) IAM Role ARN
- `managed_streams` (Boolean) Let Polytomic manage DynamoDB Stream settings
- `secret_access_key` (String, Sensitive) AWS Secret Access Key

#### Read-Only

- `aws_user` (String) User ARN
- `external_id` (String) External ID

    External ID for the IAM role


