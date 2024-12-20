---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_zendesk_chat_connection Data Source - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Zendesk Chat Connection
---

# polytomic_zendesk_chat_connection (Data Source)

Zendesk Chat Connection

## Example Usage

```terraform
data "polytomic_zendesk_chat_connection" "zendesk_chat" {
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

- `custom_api_limits` (Boolean)
- `domain` (String)
- `ratelimit_rpm` (Number) Set a custom maximum request per minute limit.

