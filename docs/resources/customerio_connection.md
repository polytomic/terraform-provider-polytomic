---
page_title: "polytomic_customerio_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Customer.io Connection
---

# polytomic_customerio_connection (Resource)

Customer.io Connection

For detailed configuration guidance, see the [Customer.io connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/customerio).

## Example Usage

```terraform
resource "polytomic_customerio_connection" "customerio" {
  name = "example"
  configuration = {
    app_api_key      = "7f81wos4eeua0kap5dabridfwgd1uh"
    site_id          = "osh853chlrhick8w7qyz"
    tracking_api_key = "a7l1nkxqhnwa48h94vbmsly1a2pnss"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Customer.io Connection identifier.
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

- `app_api_key` (String, Sensitive, Required) App API Key
- `region` (String, Optional) Datacenter region we detected
- `site_id` (String, Required) Site ID
- `tracking_api_key` (String, Sensitive, Required) Tracking API Key

