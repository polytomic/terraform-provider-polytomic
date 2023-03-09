---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_ascend_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Ascend Connection
---

# polytomic_ascend_connection (Resource)

Ascend Connection

## Example Usage

```terraform
resource "polytomic_ascend_connection" "ascend" {
  name = "example"
  configuration = {
    api_key = "my-api-key"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `configuration` (Attributes) (see [below for nested schema](#nestedatt--configuration))
- `name` (String)

### Optional

- `organization` (String) Organization ID

### Read-Only

- `id` (String) Ascend Connection identifier

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

Required:

- `api_key` (String, Sensitive)

