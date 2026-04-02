---
page_title: "polytomic_marketo_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Marketo Connection
---

# polytomic_marketo_connection (Resource)

Marketo Connection

For detailed configuration guidance, see the [Marketo connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/marketo).

## Example Usage

```terraform
resource "polytomic_marketo_connection" "marketo" {
  name = "example"
  configuration = {
    client_id     = "629b6d74-f602-47f4-8fef-388485343d85"
    client_secret = "123*******************xyz"
    rest_endpoint = "https://123-ABC-999.mktorest.com/rest"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Marketo Connection identifier.
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

- `client_id` (String, Required) Client ID
- `client_secret` (String, Sensitive, Required) Client Secret
- `concurrent_imports` (Number, Optional) Concurrent import jobs Default: <code>5</code>.
- `daily_api_calls` (Number, Optional) Daily call limit Default: <code>37500</code>.
- `enforce_api_limits` (Boolean, Optional) Enforce API limits
- `include_static_lists` (Boolean, Optional) Include static list support Default: <code>true</code>.
- `oauth_token_expiry` (String, Optional)
- `rest_endpoint` (String, Required) REST Endpoint

