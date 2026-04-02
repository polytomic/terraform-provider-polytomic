---
page_title: "polytomic_zendesk_support_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Zendesk Support Connection
---

# polytomic_zendesk_support_connection (Resource)

Zendesk Support Connection

For detailed configuration guidance, see the [Zendesk Support connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/zendesk_support).

## Example Usage

```terraform
resource "polytomic_zendesk_support_connection" "zendesk_support" {
  name = "example"
  configuration = {
    api_token           = "secret-token"
    auth_method         = "apitoken"
    domain              = "polytomic.zendesk.com"
    email               = "user@example.com"
    oauth_refresh_token = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Zendesk Support Connection identifier.
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

- `api_token` (String, Sensitive, Optional) API token
- `auth_method` (String, Required) Authentication method Valid values: <code>apitoken</code> (API token), <code>oauth</code> (OAuth). Default: <code>oauth</code>.
- `custom_api_limits` (Boolean, Optional) Enforce custom API limits
- `domain` (String, Required) Zendesk Subdomain
- `email` (String, Optional)
- `oauth_refresh_token` (String, Sensitive, Optional)
- `oauth_token_expiry` (String, Optional)
- `ratelimit_rpm` (Number, Optional) Maximum requests per minute

    Set a custom maximum request per minute limit.

