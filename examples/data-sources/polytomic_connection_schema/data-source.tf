# Example: Retrieve connection schema information
#
# This data source allows you to discover schema details including
# available fields and their current primary key configuration.

# Reference an existing connection
data "polytomic_connection_schema" "example" {
  connection_id = var.connection_id
  schema_id     = "Account" # Schema/table/object name
}

# Access the schema information
output "schema_name" {
  value = data.polytomic_connection_schema.example.name
}

output "schema_fields" {
  description = "List of all fields in the schema"
  value       = data.polytomic_connection_schema.example.fields
}

# Example: Use the data source to configure primary keys
resource "polytomic_connection_schema_primary_keys" "example_pk" {
  connection_id = data.polytomic_connection_schema.example.connection_id
  schema_id     = data.polytomic_connection_schema.example.schema_id

  # Extract field IDs from the data source
  field_ids = [
    # You can reference specific fields from the data source
    # This example shows how you might dynamically select fields
    for field in data.polytomic_connection_schema.example.fields :
    field.id if field.name == "UniqueIdentifier"
  ]
}
