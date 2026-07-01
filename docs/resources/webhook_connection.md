---
page_title: "polytomic_webhook_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Webhook Connection
---

# polytomic_webhook_connection (Resource)

Webhook Connection

For detailed configuration guidance, see the [Webhook connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/webhook).

## Example Usage

```terraform
resource "polytomic_webhook_connection" "webhook" {
  name = "example"
  configuration = {
    url = "https://example.com/webhook"
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

- `id` (String) Webhook Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `auth_method` (String) Authentication method Valid values: <code>polytomic_secret</code> (Polytomic secret), <code>basic</code> (Basic authentication), <code>header</code> (Custom header), <code>query</code> (Query string key), <code>oauth_client_credentials</code> (OAuth client credentials). Default: <code>polytomic_secret</code>.
- `url` (String) Webhook URL

#### Optional

- `basic` (Attributes) Basic authentication See [below for nested schema](#nestedatt--configuration--basic).
- `header` (Attributes) See [below for nested schema](#nestedatt--configuration--header).
- `headers` (Attributes Set) Additional headers See [below for nested schema](#nestedatt--configuration--headers).
- `oauth` (Attributes) OAuth client credentials See [below for nested schema](#nestedatt--configuration--oauth).
- `query` (Attributes Set) Query string authentication parameters See [below for nested schema](#nestedatt--configuration--query).

#### Read-Only

- `secret` (String, Sensitive)


<a id="nestedatt--configuration--basic"></a>
### Nested Schema for `configuration.basic`

#### Optional

- `password` (String)
- `username` (String)


<a id="nestedatt--configuration--header"></a>
### Nested Schema for `configuration.header`

#### Optional

- `name` (String)
- `value` (String)


<a id="nestedatt--configuration--headers"></a>
### Nested Schema for `configuration.headers`

#### Optional

- `name` (String)
- `value` (String)


<a id="nestedatt--configuration--oauth"></a>
### Nested Schema for `configuration.oauth`

#### Optional

- `auth_style` (Number) Auth style
- `client_id` (String) Client ID
- `client_secret` (String) Client secret
- `extra_form_data` (Attributes Set) Extra form data See [below for nested schema](#nestedatt--configuration--oauth--extra_form_data).
- `scopes` (Set of String)
- `token_endpoint` (String) Token endpoint


<a id="nestedatt--configuration--oauth--extra_form_data"></a>
### Nested Schema for `configuration.oauth.extra_form_data`

#### Optional

- `name` (String)
- `value` (String)


<a id="nestedatt--configuration--query"></a>
### Nested Schema for `configuration.query`

#### Optional

- `name` (String)
- `value` (String)


