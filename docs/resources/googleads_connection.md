---
page_title: "polytomic_googleads_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Google Ads Connection
---

# polytomic_googleads_connection (Resource)

Google Ads Connection

For detailed configuration guidance, see the [Google Ads connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/googleads).

## Example Usage

```terraform
resource "polytomic_googleads_connection" "googleads" {
  name = "example"
  configuration = {
    client_id           = "a45gadsfdsaf47byor2ugfbhsgllpf12gf56gfds"
    client_secret       = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    oauth_refresh_token = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Google Ads Connection identifier.
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

- `accounts` (Attributes Set, Optional) See [below for nested schema](#nestedatt--configuration--accounts).
- `blanket_user_consent` (Boolean, Optional) All transmitted users consented to ad personalization and information sharing with Google Ads

    Causes this connection to send signals to Google Ads indicating that every transmitted user has accepted ad personalization and data sharing policies. This will cause the user to be included in more advertising functions
- `client_id` (String, Sensitive, Optional)
- `client_secret` (String, Sensitive, Optional)
- `connected_user` (String, Optional) Connected user's email
- `custom_reports` (String, Optional) Custom reports

    One report per line. Format is a report name:ads object:field list. e.g. myReport:ad_groups:campaign.id
- `oauth_refresh_token` (String, Sensitive, Optional)
- `oauth_token_expiry` (String, Optional)

<a id="nestedatt--configuration--accounts"></a>
### Nested Schema for `configuration.accounts`

- `label` (String, Optional)
- `value` (String, Optional)

