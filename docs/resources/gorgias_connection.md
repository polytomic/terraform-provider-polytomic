---
page_title: "polytomic_gorgias_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Gorgias Connection
---

# polytomic_gorgias_connection (Resource)

Gorgias Connection

For detailed configuration guidance, see the [Gorgias connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/gorgias).

## Example Usage

```terraform
resource "polytomic_gorgias_connection" "gorgias" {
  name = "example"
  configuration = {
    apikey = "secret-key"
    domain = "acme"
    email  = "user@example.com"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Gorgias Connection identifier.
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

- `apikey` (String, Sensitive, Required) API Key
- `domain` (String, Required) Your Gorgias subdomain (e.g. 'acme' for acme.gorgias.com)
- `email` (String, Required) Your Gorgias account email address

