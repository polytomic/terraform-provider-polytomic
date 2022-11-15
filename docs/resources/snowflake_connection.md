---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "polytomic_snowflake_connection Resource - terraform-provider-polytomic"
subcategory: ""
description: |-
  Snowflake Connection
---

# polytomic_snowflake_connection (Resource)

Snowflake Connection

## Example Usage

```terraform
resource "polytomic_snowflake_connection" "snowflake" {
  organization = polytomic_organization.acme.id
  name         = "example"
  configuration = {
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `configuration` (Attributes) (see [below for nested schema](#nestedatt--configuration))
- `name` (String)
- `organization` (String) Organization ID

### Read-Only

- `id` (String) Snowflake Connection identifier

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

Required:

- `account` (String)
- `database` (String)
- `password` (String, Sensitive)
- `username` (String)
- `warehouse` (String)

Optional:

- `additional_params` (String)

