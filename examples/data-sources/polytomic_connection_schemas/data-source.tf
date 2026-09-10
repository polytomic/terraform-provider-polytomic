# Example: List the schemas on a connection
#
# Fields are omitted unless include_fields is true, which keeps the response
# small on connections with many schemas.

data "polytomic_connection_schemas" "warehouse" {
  connection_id  = var.connection_id
  include_fields = true
}

output "schema_ids" {
  value = [for s in data.polytomic_connection_schemas.warehouse.schemas : s.id]
}

# Schemas with no primary key, including overrides
output "schemas_without_primary_key" {
  value = [
    for s in data.polytomic_connection_schemas.warehouse.schemas :
    s.id if length([for f in s.fields : f.id if f.is_primary_key]) == 0
  ]
}
