---
page_title: "polytomic_mssql_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Microsoft SQL Server Connection
---

# polytomic_mssql_connection (Resource)

Microsoft SQL Server Connection

For detailed configuration guidance, see the [Microsoft SQL Server connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/mssql).

## Example Usage

```terraform
resource "polytomic_mssql_connection" "mssql" {
  name = "example"
  configuration = {
    database = "sampledb"
    hostname = "example.database.windows.net"
    password = "secret"
    ssh_host = "bastion.example.com"
    username = "user"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Microsoft SQL Server Connection identifier.
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

- `change_detection` (Boolean, Optional) Use change data capture for bulk syncs Default: <code>false</code>.
- `database` (String, Required)
- `hostname` (String, Required) Server
- `password` (String, Sensitive, Required)
- `port` (Number, Required) Default: <code>1433</code>.
- `ssh` (Boolean, Optional) Connect over SSH tunnel
- `ssh_host` (String, Optional) SSH host
- `ssh_port` (Number, Optional) SSH port Default: <code>22</code>.
- `ssh_private_key` (String, Sensitive, Optional) Private key
- `ssh_user` (String, Optional) SSH user Default: <code>root</code>.
- `ssl` (Boolean, Optional) Use SSL Default: <code>true</code>.
- `username` (String, Required)

