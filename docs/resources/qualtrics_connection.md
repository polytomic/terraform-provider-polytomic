---
page_title: "polytomic_qualtrics_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Qualtrics Connection
---

# polytomic_qualtrics_connection (Resource)

Qualtrics Connection

For detailed configuration guidance, see the [Qualtrics connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/qualtrics).

## Example Usage

```terraform
resource "polytomic_qualtrics_connection" "qualtrics" {
  name = "example"
  configuration = {
    api_key = "secret"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Qualtrics Connection identifier.
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

- `api_key` (String, Sensitive, Required) API Token
- `data_center` (String, Required) Data Center Valid values: <code>portland</code> (Portland, Oregon, USA), <code>washington_dc</code> (Washington, DC, USA), <code>arizona</code> (Arizona, USA (az1)), <code>us_government</code> (US Government), <code>canada</code> (Canada), <code>eu</code> (EU), <code>london</code> (London, UK), <code>singapore</code> (Singapore), <code>sydney</code> (Sydney, Australia), <code>tokyo</code> (Tokyo, Japan). Default: <code>portland</code>.

