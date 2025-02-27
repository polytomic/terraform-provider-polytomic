---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_api_connection Data Source - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  HTTP API Connection
---

# polytomic_api_connection (Data Source)

HTTP API Connection

## Example Usage

```terraform
data "polytomic_api_connection" "api" {
  id = "aab123aa-27f3-abc1-9999-abcde123a4aa"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `configuration` (Attributes) (see [below for nested schema](#nestedatt--configuration))
- `organization` (String)

### Read-Only

- `id` (String) The ID of this resource.
- `name` (String)

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

Read-Only:

- `auth` (Attributes) Authentication method (see [below for nested schema](#nestedatt--configuration--auth))
- `body` (String) JSON payload
- `headers` (Attributes Set) (see [below for nested schema](#nestedatt--configuration--headers))
- `healthcheck` (String) Health check endpoint

    Path to request when checking the health of this connection. No health check will be performed if left empty.
- `parameters` (Attributes Set) Query string parameters (see [below for nested schema](#nestedatt--configuration--parameters))
- `url` (String) Base URL

<a id="nestedatt--configuration--auth"></a>
### Nested Schema for `configuration.auth`

Read-Only:

- `basic` (Attributes) Basic authentication (see [below for nested schema](#nestedatt--configuration--auth--basic))
- `header` (Attributes) Header key (see [below for nested schema](#nestedatt--configuration--auth--header))
- `oauth` (Attributes) (see [below for nested schema](#nestedatt--configuration--auth--oauth))

<a id="nestedatt--configuration--auth--basic"></a>
### Nested Schema for `configuration.auth.basic`

Read-Only:

- `password` (String)
- `username` (String)


<a id="nestedatt--configuration--auth--header"></a>
### Nested Schema for `configuration.auth.header`

Read-Only:

- `name` (String)
- `value` (String)


<a id="nestedatt--configuration--auth--oauth"></a>
### Nested Schema for `configuration.auth.oauth`

Read-Only:

- `auth_style` (Number) Auth style
- `client_id` (String) Client ID
- `client_secret` (String) Client secret
- `extra_form_data` (Attributes Set) Extra form data (see [below for nested schema](#nestedatt--configuration--auth--oauth--extra_form_data))
- `scopes` (Set of String)
- `token_endpoint` (String) Token endpoint

<a id="nestedatt--configuration--auth--oauth--extra_form_data"></a>
### Nested Schema for `configuration.auth.oauth.token_endpoint`

Read-Only:

- `name` (String)
- `value` (String)




<a id="nestedatt--configuration--headers"></a>
### Nested Schema for `configuration.headers`

Read-Only:

- `name` (String)
- `value` (String)


<a id="nestedatt--configuration--parameters"></a>
### Nested Schema for `configuration.parameters`

Read-Only:

- `name` (String)
- `value` (String)


