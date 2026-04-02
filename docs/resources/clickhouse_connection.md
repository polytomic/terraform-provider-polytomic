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
    database = "default"
    hostname = "clickhouse.example.com"
    ssh_host = "bastion.example.com"
    username = "default"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) ClickHouse Connection identifier.
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

- `database` (String, Optional)
- `hostname` (String, Required)
- `password` (String, Sensitive, Optional)
- `port` (Number, Required) Default: <code>9440</code>.
- `skip_verify` (Boolean, Optional) Skip certificate verification Default: <code>true</code>.
- `ssh` (Boolean, Optional) Connect over SSH tunnel
- `ssh_host` (String, Optional) SSH host
- `ssh_port` (Number, Optional) SSH port Default: <code>22</code>.
- `ssh_private_key` (String, Sensitive, Optional) Private key
- `ssh_user` (String, Optional) SSH user Default: <code>root</code>.
- `ssl` (Boolean, Optional) Use SSL Default: <code>true</code>.
- `username` (String, Required)

