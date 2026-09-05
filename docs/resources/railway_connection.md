---
page_title: "polytomic_railway_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Railway Connection
---

# polytomic_railway_connection (Resource)

Railway Connection

For detailed configuration guidance, see the [Railway connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/railway).

## Example Usage

```terraform
resource "polytomic_railway_connection" "railway" {
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

- `id` (String) Railway Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `api_key` (String, Sensitive) API key


