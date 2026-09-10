# Example: Retrieve connection schema information
#
# This data source reports a schema's fields, including their primary key
# status and any user-defined overrides.

data "polytomic_connection_schema" "example" {
  connection_id = var.connection_id
  schema_id     = "Account" # Schema/table/object name
}

output "schema_name" {
  value = data.polytomic_connection_schema.example.name
}

# Look up a single field by ID
output "id_field_type" {
  value = data.polytomic_connection_schema.example.fields_by_id["Id"].type
}

# The schema's primary key, including any override
output "primary_key" {
  value = [
    for field in data.polytomic_connection_schema.example.fields :
    field.id if field.is_primary_key
  ]
}

# The primary key the source reports, ignoring overrides
output "source_primary_key" {
  value = [
    for field in data.polytomic_connection_schema.example.fields :
    field.id if coalesce(field.source_primary_key, false)
  ]
}
