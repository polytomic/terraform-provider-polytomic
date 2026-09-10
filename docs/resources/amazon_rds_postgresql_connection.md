---
page_title: "polytomic_amazon_rds_postgresql_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Amazon RDS PostgreSQL Connection
---

# polytomic_amazon_rds_postgresql_connection (Resource)

Amazon RDS PostgreSQL Connection

For detailed configuration guidance, see the [Amazon RDS PostgreSQL connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/amazon_rds_postgresql).

## Example Usage

```terraform
resource "polytomic_amazon_rds_postgresql_connection" "amazon_rds_postgresql" {
  name = "example"
  configuration = {
    database     = "sampledb"
    hostname     = "database.123456789012.us-east-1.rds.amazonaws.com"
    iam_role_arn = "arn:aws:iam::123456789012:role/polytomic-rds"
    publication  = "polytomic"
    region       = "us-east-1"
    ssh_host     = "bastion.example.com"
    username     = "polytomic"
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

- `id` (String) Amazon RDS PostgreSQL Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `database` (String)
- `hostname` (String) RDS endpoint

    The Amazon RDS endpoint; custom DNS names cannot be used to generate IAM authentication tokens
- `iam_role_arn` (String) IAM role ARN

    Role that grants rds-db:connect access to this database user
- `port` (Number) Default: <code>5432</code>.
- `region` (String) AWS region
- `username` (String)

#### Optional

- `change_detection` (Boolean) Use logical replication for bulk syncs Default: <code>false</code>.
- `publication` (String)
- `ssh` (Boolean) Connect over SSH tunnel
- `ssh_host` (String) SSH host
- `ssh_port` (Number) SSH port Default: <code>22</code>.
- `ssh_private_key` (String, Sensitive) Private key
- `ssh_user` (String) SSH user Default: <code>root</code>.
- `tags` (Map of String) Additional tags to apply during role assumption

#### Read-Only

- `external_id` (String) External ID

    External ID for the IAM role


