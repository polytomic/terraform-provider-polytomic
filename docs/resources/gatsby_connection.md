---
page_title: "polytomic_gatsby_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Gatsby Connection
---

# polytomic_gatsby_connection (Resource)

Gatsby Connection

For detailed configuration guidance, see the [Gatsby connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/gatsby).

## Example Usage

```terraform
resource "polytomic_gatsby_connection" "gatsby" {
  name = "example"
  configuration = {
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

- `id` (String) Gatsby Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `email` (String)
- `password` (String, Sensitive)

#### Optional

- `organizations` (Attributes Set) See [below for nested schema](#nestedatt--configuration--organizations).


<a id="nestedatt--configuration--organizations"></a>
### Nested Schema for `configuration.organizations`

#### Optional

- `label` (String)
- `value` (String)


