---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_postgresql_connection Data Source - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  PostgreSQL Connection
---

# polytomic_postgresql_connection (Data Source)

PostgreSQL Connection

## Example Usage

```terraform
data "polytomic_postgresql_connection" "postgresql" {
  id = "aab123aa-27f3-abc1-9999-abcde123a4aa"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `configuration` (Attributes) (see [below for nested schema](#nestedatt--configuration))
- `organization` (String)

### Read-Only

- `id` (String) The ID of this resource.
- `name` (String)

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

Read-Only:

- `change_detection` (Boolean) Use logical replication for bulk syncs
- `client_certs` (Boolean) Use client certificates
- `database` (String)
- `hostname` (String)
- `port` (Number)
- `publication` (String)
- `ssh` (Boolean) Connect over SSH tunnel
- `ssh_host` (String) SSH host
- `ssh_port` (Number) SSH port
- `ssh_user` (String) SSH user
- `ssl` (Boolean) Use SSL
- `username` (String)


