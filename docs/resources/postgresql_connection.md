---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_postgresql_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  PostgreSQL Connection
---

# polytomic_postgresql_connection (Resource)

PostgreSQL Connection

## Example Usage

```terraform
resource "polytomic_postgresql_connection" "postgresql" {
  name = "example"
  configuration = {
    database    = "sampledb"
    hostname    = "database.example.com"
    password    = "password"
    publication = "polytomic"
    ssh_host    = "bastion.example.com"
    username    = "postgres"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `configuration` (Attributes) (see [below for nested schema](#nestedatt--configuration))
- `name` (String)

### Optional

- `force_destroy` (Boolean) Indicates whether dependent models, syncs, and bulk syncs should be cascade
deleted when this connection is destroy.

  This only deletes other resources when the connection is destroyed, not when
setting this parameter to `true`. Once this parameter is set to `true`, there
must be a successful `terraform apply` run before a destroy is required to
update this value in the resource state. Without a successful `terraform apply`
after this parameter is set, this flag will have no effect. If setting this
field in the same operation that would require replacing the connection or
destroying the connection, this flag will not work. Additionally when importing
a connection, a successful `terraform apply` is required to set this value in
state before it will take effect on a destroy operation.
- `organization` (String) Organization ID

### Read-Only

- `id` (String) PostgreSQL Connection identifier

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

Required:

- `database` (String)
- `hostname` (String)
- `password` (String, Sensitive)
- `port` (Number)
- `username` (String)

Optional:

- `ca_cert` (String, Sensitive) CA certificate
- `change_detection` (Boolean) Use logical replication for bulk syncs
- `client_certificate` (String, Sensitive) Client certificate
- `client_certs` (Boolean) Use client certificates
- `client_key` (String, Sensitive) Client key
- `publication` (String)
- `ssh` (Boolean) Connect over SSH tunnel
- `ssh_host` (String) SSH host
- `ssh_port` (Number) SSH port
- `ssh_private_key` (String, Sensitive) Private key
- `ssh_user` (String) SSH user
- `ssl` (Boolean) Use SSL


