---
page_title: "polytomic_x_ads_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  X Ads Connection
---

# polytomic_x_ads_connection (Resource)

X Ads Connection

For detailed configuration guidance, see the [X Ads connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/x_ads).

## Example Usage

```terraform
resource "polytomic_x_ads_connection" "x_ads" {
  name = "example"
  configuration = {
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

- `id` (String) X Ads Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `access_token` (String, Sensitive) Access token
- `access_token_secret` (String, Sensitive) Access token secret
- `consumer_key` (String, Sensitive) Consumer key
- `consumer_secret` (String, Sensitive) Consumer secret


