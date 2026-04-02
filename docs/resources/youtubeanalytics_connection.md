---
page_title: "polytomic_youtubeanalytics_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  YouTube Analytics Connection
---

# polytomic_youtubeanalytics_connection (Resource)

YouTube Analytics Connection

For detailed configuration guidance, see the [YouTube Analytics connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/youtubeanalytics).

## Example Usage

```terraform
resource "polytomic_youtubeanalytics_connection" "youtubeanalytics" {
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
- `id` (String, Read-only) YouTube Analytics Connection identifier.
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
- `content_owner_id` (String, Optional) Content owner ID

    If you are using a content owner account enter the content owner ID here. This is required for some reports.
- `oauth_refresh_token` (String, Sensitive, Optional)
- `oauth_token_expiry` (String, Optional)
- `user_email` (String, Optional) Connected user's email

