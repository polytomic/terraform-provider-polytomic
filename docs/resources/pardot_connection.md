---
page_title: "polytomic_pardot_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Pardot Connection
---

# polytomic_pardot_connection (Resource)

Pardot Connection

For detailed configuration guidance, see the [Pardot connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/pardot).

## Example Usage

```terraform
resource "polytomic_pardot_connection" "pardot" {
  name = "example"
  configuration = {
    account_type     = "Production"
    business_unit_id = "1234567"
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

- `id` (String) Pardot Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Optional

- `account_type` (String) Account type Valid values: <code>Production</code>, <code>Sandbox</code>, <code>Demo</code>.
- `business_unit_id` (String) Business Unit ID
- `daily_api_calls` (Number) Daily call limit
- `enforce_api_limits` (Boolean) Enforce API limits

#### Read-Only

- `username` (String)


