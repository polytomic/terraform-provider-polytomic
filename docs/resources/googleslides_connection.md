---
page_title: "polytomic_googleslides_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Google Slides Connection
---

# polytomic_googleslides_connection (Resource)

Google Slides Connection

For detailed configuration guidance, see the [Google Slides connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/googleslides).

## Example Usage

```terraform
resource "polytomic_googleslides_connection" "googleslides" {
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
- `id` (String, Read-only) Google Slides Connection identifier.
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
- `folder_id` (Attributes, Required) Folder See [below for nested schema](#nestedatt--configuration--folder_id).
- `include_subdirectories` (Boolean, Optional) Include Subdirectories Default: <code>false</code>.
- `oauth_refresh_token` (String, Sensitive, Optional)
- `oauth_token_expiry` (String, Optional)
- `service_account` (String, Sensitive, Optional) Service account key
- `user_email` (String, Optional) Connected user's email

<a id="nestedatt--configuration--folder_id"></a>
### Nested Schema for `configuration.folder_id`

- `label` (String, Optional)
- `value` (String, Optional)

