---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_shortio_connection Data Source - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Short.io Connection
---

# polytomic_shortio_connection (Data Source)

Short.io Connection

## Example Usage

```terraform
data "polytomic_shortio_connection" "shortio" {
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

