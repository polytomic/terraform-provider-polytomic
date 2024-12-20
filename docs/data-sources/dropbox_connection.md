---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_dropbox_connection Data Source - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Dropbox Connection
---

# polytomic_dropbox_connection (Data Source)

Dropbox Connection

## Example Usage

```terraform
data "polytomic_dropbox_connection" "dropbox" {
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

- `bucket` (String)
- `is_single_table` (Boolean) Treat the files as a single table.
- `oauth_token_expiry` (String)
- `single_table_file_format` (String)
- `single_table_name` (String)
- `skip_lines` (Number) Skip first N lines of each CSV file.

