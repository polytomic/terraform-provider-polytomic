---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_mysql_connection Resource - terraform-provider-polytomic"
subcategory: "Connection"
description: |-
  MySQL Connection
---

# polytomic_mysql_connection (Resource)

MySQL Connection

## Example Usage

```terraform
resource "polytomic_mysql_connection" "mysql" {
  name = "example"
  configuration = {
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

- `id` (String) MySQL Connection identifier

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

Required:

- `account` (String)
- `dbname` (String)
- `hostname` (String)
- `passwd` (String, Sensitive)

Optional:

- `port` (Number)
- `private_key` (String, Sensitive)
- `ssh` (Boolean)
- `ssh_host` (String)
- `ssh_port` (Number)
- `ssh_user` (String)

