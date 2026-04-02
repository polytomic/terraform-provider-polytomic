---
page_title: "polytomic_jira_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Jira Connection
---

# polytomic_jira_connection (Resource)

Jira Connection

For detailed configuration guidance, see the [Jira connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/jira).

## Example Usage

```terraform
resource "polytomic_jira_connection" "jira" {
  name = "example"
  configuration = {
    access_token = "secret"
    auth_method  = "pat"
    url          = "https://jira.mycompany.com/"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Jira Connection identifier.
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

- `access_token` (String, Sensitive, Optional) Personal access token
- `api_key` (String, Sensitive, Optional) API token
- `auth_method` (String, Required) Authentication method Valid values: <code>apikey</code> (API token), <code>pat</code> (Personal access token). Default: <code>apikey</code>.
- `url` (String, Required) Jira URL
- `username` (String, Optional)

