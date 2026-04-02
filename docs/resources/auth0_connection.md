---
page_title: "polytomic_auth0_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Auth0 Connection
---

# polytomic_auth0_connection (Resource)

Auth0 Connection

For detailed configuration guidance, see the [Auth0 connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/auth0).

## Example Usage

```terraform
resource "polytomic_auth0_connection" "auth0" {
  name = "example"
  configuration = {
    client_id     = "jI2Zem1Yzxy8s8s..."
    client_secret = "cB6NNPhR12R8pJ7M..."
    domain        = "dev-g1ce1rt9.us.auth0.com"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Auth0 Connection identifier.
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

- `client_id` (String, Required) App Client ID
- `client_secret` (String, Sensitive, Required) App Client Secret
- `domain` (String, Required) The domain of the Auth0 instance

