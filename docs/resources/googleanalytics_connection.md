---
page_title: "polytomic_googleanalytics_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Google Analytics Connection
---

# polytomic_googleanalytics_connection (Resource)

Google Analytics Connection

For detailed configuration guidance, see the [Google Analytics connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/googleanalytics).

## Example Usage

```terraform
resource "polytomic_googleanalytics_connection" "googleanalytics" {
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
- `id` (String, Read-only) Google Analytics Connection identifier.
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

- `auth_method` (String, Required) Authentication method Valid values: <code>oauth</code> (OAuth), <code>jwt</code> (Service Account). Default: <code>oauth</code>.
- `client_id` (String, Sensitive, Optional)
- `client_secret` (String, Sensitive, Optional)
- `custom_reports` (String, Optional) Custom reports

    One report per line. Format is a report name followed by a comma-separated list of fields. e.g. myReport:field1
- `oauth_refresh_token` (String, Sensitive, Optional)
- `oauth_token_expiry` (String, Optional)
- `properties` (Attributes Set, Optional) See [below for nested schema](#nestedatt--configuration--properties).
- `service_account` (String, Sensitive, Optional) Service account key
- `user_email` (String, Optional) Connected user's email

<a id="nestedatt--configuration--properties"></a>
### Nested Schema for `configuration.properties`

- `label` (String, Optional)
- `value` (String, Optional)

