---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_gong_connection Data Source - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Gong Connection
---

# polytomic_gong_connection (Data Source)

Gong Connection

## Example Usage

```terraform
data "polytomic_gong_connection" "gong" {
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

- `access_key` (String)
- `auth_method` (String)
- `oauth_token_expiry` (String)
- `subdomain` (String) Gong subdomain i.e. company-17 if you access Gong via https://company-17.app.gong.io

