---
page_title: "polytomic_scylladb_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  ScyllaDB Connection
---

# polytomic_scylladb_connection (Resource)

ScyllaDB Connection

For detailed configuration guidance, see the [ScyllaDB connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/scylladb).

## Example Usage

```terraform
resource "polytomic_scylladb_connection" "scylladb" {
  name = "example"
  configuration = {
    hosts    = "scylla.example.com"
    password = "password"
    ssh_host = "bastion.example.com"
    username = "scylla"
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

- `id` (String) ScyllaDB Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `hosts` (String) Hostname(s)

    Comma-separated list

#### Optional

- `ca_cert` (String) CA certificate
- `client_certificate` (String, Sensitive) Client certificate
- `client_certs` (Boolean) Use client certificates Default: <code>false</code>.
- `client_key` (String, Sensitive) Client key
- `password` (String, Sensitive)
- `skip_verify` (Boolean) Skip certificate verification Default: <code>false</code>.
- `ssh` (Boolean) Connect over SSH tunnel
- `ssh_host` (String) SSH host
- `ssh_port` (Number) SSH port Default: <code>22</code>.
- `ssh_private_key` (String, Sensitive) Private key
- `ssh_user` (String) SSH user Default: <code>root</code>.
- `tls` (Boolean) Use TLS/SSL Default: <code>false</code>.
- `username` (String)


