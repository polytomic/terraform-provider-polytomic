---
page_title: "polytomic_ibm_db2_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  IBM Db2 Connection
---

# polytomic_ibm_db2_connection (Resource)

IBM Db2 Connection

For detailed configuration guidance, see the [IBM Db2 connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/ibm_db2).

## Example Usage

```terraform
resource "polytomic_ibm_db2_connection" "ibm_db2" {
  name = "example"
  configuration = {
    account  = "db2admin"
    database = "SAMPLE"
    hostname = "db2.example.com"
    passwd   = "password"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) IBM Db2 Connection identifier.
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

- `account` (String, Required) Username
- `database` (String, Required)
- `hostname` (String, Required)
- `passwd` (String, Sensitive, Required) Password
- `ssl` (Boolean, Optional) Use SSL Default: <code>true</code>.

