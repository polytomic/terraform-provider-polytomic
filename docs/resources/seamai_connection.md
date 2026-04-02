---
page_title: "polytomic_seamai_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Seam AI Connection
---

# polytomic_seamai_connection (Resource)

Seam AI Connection

For detailed configuration guidance, see the [Seam AI connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/seamai).

## Example Usage

```terraform
resource "polytomic_seamai_connection" "seamai" {
  name = "example"
  configuration = {
    apikey_id     = "9snbax8ij4hvAvBv3ap3EpQ"
    apikey_secret = "9snbax8ij4hvAvBv3ap3EpQ"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Seam AI Connection identifier.
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

- `apikey_id` (String, Sensitive, Required) API Key ID
- `apikey_secret` (String, Sensitive, Required) API Key Secret
- `base_url` (String, Optional) Alternative base URL

    Alternate environment API URL (including any necessary paths

