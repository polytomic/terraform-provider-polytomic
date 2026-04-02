---
page_title: "polytomic_dialpad_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Dialpad Connection
---

# polytomic_dialpad_connection (Resource)

Dialpad Connection

For detailed configuration guidance, see the [Dialpad connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/dialpad).

## Example Usage

```terraform
resource "polytomic_dialpad_connection" "dialpad" {
  name = "example"
  configuration = {
    api_key             = "2px8lqeoakuea2ccl4rxm13i6tb"
    application_id      = "a45gadsfdsaf47byor2ugfbhsgllpf12gf56gfds"
    auth_method         = "api_key"
    client_secret       = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    oauth_refresh_token = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Dialpad Connection identifier.
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
- `auth_method` (String, Required) Authentication method Valid values: <code>api_key</code> (API Key), <code>oauth</code> (OAuth). Default: <code>api_key</code>.
- `client_secret` (String, Sensitive, Optional)
- `oauth_refresh_token` (String, Sensitive, Optional)
- `oauth_token_expiry` (String, Optional)

