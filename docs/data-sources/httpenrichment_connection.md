---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_httpenrichment_connection Data Source - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  HTTP Enrichment Connection
---

# polytomic_httpenrichment_connection (Data Source)

HTTP Enrichment Connection

## Example Usage

```terraform
data "polytomic_httpenrichment_connection" "httpenrichment" {
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

- `auth` (Attributes) (see [below for nested schema](#nestedatt--configuration--auth))
- `body` (String)
- `fields` (String) List of fields to be returned by the enrichment
- `headers` (String)
- `healthcheck` (String) Path to request when checking the health of this connection. No health check will be performed if left empty.
- `input_mappings` (String) List of input mappings to be used in the query. Each mapping should be a valid JSONPath expression.
- `method` (String)
- `parameters` (String)
- `url` (String)

<a id="nestedatt--configuration--auth"></a>
### Nested Schema for `configuration.auth`

Read-Only:

- `basic` (Attributes) (see [below for nested schema](#nestedatt--configuration--auth--basic))
- `header` (Attributes) (see [below for nested schema](#nestedatt--configuration--auth--header))
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

- `auth_style` (Number)
- `client_id` (String)
- `client_secret` (String)
- `extra_form_data` (String)
- `scopes` (String)
- `token_endpoint` (String)


