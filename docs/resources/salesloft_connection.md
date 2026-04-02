---
page_title: "polytomic_salesloft_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Salesloft Connection
---

# polytomic_salesloft_connection (Resource)

Salesloft Connection

For detailed configuration guidance, see the [Salesloft connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/salesloft).

## Example Usage

```terraform
resource "polytomic_salesloft_connection" "salesloft" {
  name = "example"
  configuration = {
    application_id      = "a45gadsfdsaf47byor2ugfbhsgllpf12gf56gfds"
    client_secret       = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    oauth_refresh_token = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Salesloft Connection identifier.
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

- `api_key` (String, Sensitive, Optional) API Key
- `application_id` (String, Sensitive, Optional)
- `auth_method` (String, Required) Authentication method Valid values: <code>oauth</code> (OAuth), <code>api_key</code> (API Key). Default: <code>oauth</code>.
- `client_secret` (String, Sensitive, Optional)
- `connected_user` (String, Optional) Connected user's email
- `oauth_refresh_token` (String, Sensitive, Optional)
- `oauth_token_expiry` (String, Optional)

