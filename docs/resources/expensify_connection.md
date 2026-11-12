---
page_title: "polytomic_expensify_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Expensify Connection
---

# polytomic_expensify_connection (Resource)

Expensify Connection

For detailed configuration guidance, see the [Expensify connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/expensify).

## Example Usage

```terraform
resource "polytomic_expensify_connection" "expensify" {
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

- `id` (String) Expensify Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `partner_user_id` (String) Partner user ID

    The Expensify partner user ID used to authenticate requests.
- `partner_user_secret` (String, Sensitive) Partner user secret

    The Expensify partner user secret paired with the partner user ID.


