---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_seamai_connection Data Source - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Seam AI Connection
---

# polytomic_seamai_connection (Data Source)

Seam AI Connection

## Example Usage

```terraform
data "polytomic_seamai_connection" "seamai" {
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

Read-Only:

- `base_url` (String) Alternative base URL

    Alternate environment API URL (including any necessary paths


