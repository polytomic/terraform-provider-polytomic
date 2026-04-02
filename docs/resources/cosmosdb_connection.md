---
page_title: "polytomic_cosmosdb_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Azure Cosmos DB Connection
---

# polytomic_cosmosdb_connection (Resource)

Azure Cosmos DB Connection

For detailed configuration guidance, see the [Azure Cosmos DB connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/cosmosdb).

## Example Usage

```terraform
resource "polytomic_cosmosdb_connection" "cosmosdb" {
  name = "example"
  configuration = {
    key = "dasfdasz62px8lqeoakuea2ccl4rxmhu1i20kc8ruvksmzxq=="
    uri = "https://contosomarketing.documents.azure.com"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Azure Cosmos DB Connection identifier.
- `force_destroy` (Boolean, Optional) Indicates whether dependent models, syncs, and bulk syncs should be
cascade-deleted when this connection is destroyed.

    This only deletes other resources when the connection is destroyed, not when
setting this parameter to `true`. Once this parameter is set to `true`, there
must be a successful `terraform apply` run before a destroy is required to
update this value in the resource state. Without a successful `terraform apply`
after this parameter is set, this flag will have no effect. If setting this
field in the same operation that would require replacing the connection or
destroying the connection, this flag will not work. Additionally when importing
a connection, a successful `terraform apply` is required to set this value in
state before it will take effect on a destroy operation.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

- `key` (String, Sensitive, Required)
- `uri` (String, Required)

