---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_snowflake_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Snowflake Connection
---

# polytomic_snowflake_connection (Resource)

Snowflake Connection

## Example Usage

```terraform
resource "polytomic_snowflake_connection" "snowflake" {
  name = "example"
  configuration = {
    account   = "uc193736182"
    dbname    = "database_name"
    password  = "password"
    username  = "user"
    warehouse = "warehouse"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `configuration` (Attributes) (see [below for nested schema](#nestedatt--configuration))
- `name` (String)

### Optional

- `force_destroy` (Boolean) Indicates whether dependent models, syncs, and bulk syncs should be cascade
deleted when this connection is destroy.

  This only deletes other resources when the connection is destroyed, not when
setting this parameter to `true`. Once this parameter is set to `true`, there
must be a successful `terraform apply` run before a destroy is required to
update this value in the resource state. Without a successful `terraform apply`
after this parameter is set, this flag will have no effect. If setting this
field in the same operation that would require replacing the connection or
destroying the connection, this flag will not work. Additionally when importing
a connection, a successful `terraform apply` is required to set this value in
state before it will take effect on a destroy operation.
- `organization` (String) Organization ID

### Read-Only

- `id` (String) Snowflake Connection identifier

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

Required:

- `account` (String) Account Name

    e.g. uc193736182, ja8382948.us-central-1.gcp
- `dbname` (String) Database
- `username` (String)

Optional:

- `key_pair_auth` (Boolean) Use key pair authentication
- `params` (String) Additional parameters

    Additional connection parameters, formatted as a query string
- `password` (String, Sensitive)
- `private_key` (String, Sensitive) Private key
- `private_key_passphrase` (String, Sensitive) Private key passphrase
- `warehouse` (String) Compute warehouse


