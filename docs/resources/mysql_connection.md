---
page_title: "polytomic_mysql_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  MySQL Connection
---

# polytomic_mysql_connection (Resource)

MySQL Connection

For detailed configuration guidance, see the [MySQL connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/mysql).

## Example Usage

```terraform
resource "polytomic_mysql_connection" "mysql" {
  name = "example"
  configuration = {
    account  = "admin"
    dbname   = "mydb"
    hostname = "database.example.com"
    passwd   = "password"
    ssh_host = "bastion.example.com"
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

- `id` (String) MySQL Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `account` (String) Username
- `hostname` (String)
- `passwd` (String, Sensitive) Password
- `port` (Number) Default: <code>3306</code>.

#### Optional

- `change_detection` (Boolean) Use replication for bulk syncs Default: <code>false</code>.
- `dbname` (String) Database
- `ssh` (Boolean) Connect over SSH tunnel
- `ssh_host` (String) SSH host
- `ssh_port` (Number) SSH port Default: <code>22</code>.
- `ssh_private_key` (String, Sensitive) Private key
- `ssh_user` (String) SSH user Default: <code>root</code>.
- `ssl` (Boolean) Use SSL Default: <code>true</code>.


