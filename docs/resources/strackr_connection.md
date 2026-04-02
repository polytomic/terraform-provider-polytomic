---
page_title: "polytomic_strackr_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Strackr Connection
---

# polytomic_strackr_connection (Resource)

Strackr Connection

For detailed configuration guidance, see the [Strackr connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/strackr).

## Example Usage

```terraform
resource "polytomic_strackr_connection" "strackr" {
  name = "example"
  configuration = {
    currency_type = "USD"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Strackr Connection identifier.
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

- `api_id` (Number, Sensitive, Required) API ID
- `api_key` (String, Sensitive, Required) API Key
- `currency_type` (String, Required) Currency Type Valid values: <code>EUR</code>, <code>USD</code>, <code>CAD</code>, <code>GBP</code>, <code>RUB</code>, <code>SEK</code>, <code>AUD</code>, <code>INR</code>, <code>NOK</code>, <code>DKK</code>. Default: <code>USD</code>.
- `linkbuilder_customs_text` (String, Optional) Linkbuilder Customs Text

