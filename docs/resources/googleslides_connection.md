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

- `id` (String) Google Slides Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `folder_id` (Attributes) Folder See [below for nested schema](#nestedatt--configuration--folder_id).

#### Optional

- `client_id` (String, Sensitive)
- `client_secret` (String, Sensitive)
- `connect_mode` (String) Authentication method

    Default: browser Valid values: <code>browser</code>, <code>jwt</code>. Default: <code>browser</code>.
- `include_subdirectories` (Boolean) Include Subdirectories Default: <code>false</code>.
- `oauth_refresh_token` (String, Sensitive)
- `service_account` (String, Sensitive) Service account key

#### Read-Only

- `user_email` (String) Connected user's email


<a id="nestedatt--configuration--folder_id"></a>
### Nested Schema for `configuration.folder_id`

#### Optional

- `label` (String)
- `value` (String)


