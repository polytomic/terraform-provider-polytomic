---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_harmonic_connection Data Source - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Harmonic Connection
---

# polytomic_harmonic_connection (Data Source)

Harmonic Connection

## Example Usage

```terraform
data "polytomic_harmonic_connection" "harmonic" {
  id = "aab123aa-27f3-abc1-9999-abcde123a4aa"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `configuration` (Attributes) (see [below for nested schema](#nestedatt--configuration))
- `organization` (String)

### Read-Only

- `id` (String) The ID of this resource.
- `name` (String)

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`


