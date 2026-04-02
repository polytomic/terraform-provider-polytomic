---
page_title: "polytomic_appcues_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Appcues Connection
---

# polytomic_appcues_connection (Resource)

Appcues Connection

For detailed configuration guidance, see the [Appcues connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/appcues).

## Example Usage

```terraform
resource "polytomic_appcues_connection" "appcues" {
  name = "example"
  configuration = {
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Appcues Connection identifier.
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

- `account_id` (String, Required) Account ID
- `api_key` (String, Sensitive, Required) API Key
- `api_secret` (String, Sensitive, Required) API Secret

