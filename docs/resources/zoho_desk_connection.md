---
page_title: "polytomic_zoho_desk_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Zoho Desk Connection
---

# polytomic_zoho_desk_connection (Resource)

Zoho Desk Connection

For detailed configuration guidance, see the [Zoho Desk connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/zoho_desk).

## Example Usage

```terraform
resource "polytomic_zoho_desk_connection" "zoho_desk" {
  name = "example"
  configuration = {
    client_id           = "1000.XXXXXXXXXXXXXXXXXXXX"
    client_secret       = "xxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    oauth_refresh_token = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Zoho Desk Connection identifier.
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

- `client_id` (String, Sensitive, Optional)
- `client_secret` (String, Sensitive, Optional)
- `oauth_refresh_token` (String, Sensitive, Optional)
- `organizations` (Attributes Set, Optional) See [below for nested schema](#nestedatt--configuration--organizations).

<a id="nestedatt--configuration--organizations"></a>
### Nested Schema for `configuration.organizations`

- `label` (String, Optional)
- `value` (String, Optional)

