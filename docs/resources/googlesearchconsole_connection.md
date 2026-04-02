---
page_title: "polytomic_googlesearchconsole_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Google Search Console Connection
---

# polytomic_googlesearchconsole_connection (Resource)

Google Search Console Connection

For detailed configuration guidance, see the [Google Search Console connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/googlesearchconsole).

## Example Usage

```terraform
resource "polytomic_googlesearchconsole_connection" "googlesearchconsole" {
  name = "example"
  configuration = {
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Google Search Console Connection identifier.
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
- `oauth_token_expiry` (String, Optional)
- `sites` (Attributes Set, Optional) See [below for nested schema](#nestedatt--configuration--sites).

<a id="nestedatt--configuration--sites"></a>
### Nested Schema for `configuration.sites`

- `label` (String, Optional)
- `value` (String, Optional)

