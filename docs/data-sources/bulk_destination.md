---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_bulk_destination Data Source - terraform-provider-polytomic"
subcategory: "Bulk Syncs"
description: |-
  Bulk Destination
---

# polytomic_bulk_destination (Data Source)

Bulk Destination

## Example Usage

```terraform
data "polytomic_bulk_destination" "dest" {
  connection_id = "bbd321bb-abc1-27f3-1111-abcde123a1bb"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `connection_id` (String)

### Read-Only

- `modes` (Set of Object) (see [below for nested schema](#nestedatt--modes))
- `organization` (String)
- `required_configuration` (Set of String)

<a id="nestedatt--modes"></a>
### Nested Schema for `modes`

Read-Only:

- `description` (String)
- `id` (String)
- `label` (String)


