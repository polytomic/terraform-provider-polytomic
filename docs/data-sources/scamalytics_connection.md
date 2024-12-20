---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_scamalytics_connection Data Source - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Scamalytics Connection
---

# polytomic_scamalytics_connection (Data Source)

Scamalytics Connection

## Example Usage

```terraform
data "polytomic_scamalytics_connection" "scamalytics" {
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

- `endpoint` (String)

