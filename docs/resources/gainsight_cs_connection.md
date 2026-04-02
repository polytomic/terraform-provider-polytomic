---
page_title: "polytomic_gainsight_cs_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Gainsight CS Connection
---

# polytomic_gainsight_cs_connection (Resource)

Gainsight CS Connection

For detailed configuration guidance, see the [Gainsight CS connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/gainsight_cs).

## Example Usage

```terraform
resource "polytomic_gainsight_cs_connection" "gainsight_cs" {
  name = "example"
  configuration = {
    domain = "company.gainsightcloud.com"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Gainsight CS Connection identifier.
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

- `access_key` (String, Sensitive, Required) Access Key

    Gainsight CS API Access Key
- `domain` (String, Required) Your Gainsight CS domain

