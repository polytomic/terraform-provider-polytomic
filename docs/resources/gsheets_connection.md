---
page_title: "polytomic_gsheets_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Google Sheets Connection
---

# polytomic_gsheets_connection (Resource)

Google Sheets Connection

For detailed configuration guidance, see the [Google Sheets connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/gsheets).

## Example Usage

```terraform
resource "polytomic_gsheets_connection" "gsheets" {
  name = "example"
  configuration = {
    client_id           = "eb669428-1854-4cb1-a560-403e05b8acbf"
    client_secret       = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    oauth_refresh_token = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Google Sheets Connection identifier.
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
- `connect_mode` (String, Optional) Default: browser Valid values: <code>browser</code>, <code>jwt</code>. Default: <code>browser</code>.
- `has_headers` (Boolean, Optional) Columns have headers
- `oauth_refresh_token` (String, Sensitive, Optional)
- `oauth_token_expiry` (String, Optional)
- `service_account` (String, Sensitive, Optional) Service account key
- `spreadsheet_id` (Attributes, Required) Spreadsheet See [below for nested schema](#nestedatt--configuration--spreadsheet_id).
- `user_email` (String, Optional) Connected user's email

<a id="nestedatt--configuration--spreadsheet_id"></a>
### Nested Schema for `configuration.spreadsheet_id`

- `label` (String, Optional)
- `value` (String, Optional)

