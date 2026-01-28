# Example: Override primary keys for a connection schema
#
# This resource allows you to specify which fields should be used as primary keys
# for a connection schema, overriding the auto-detected primary keys from the source.

# First, create or reference a connection
resource "polytomic_salesforce_connection" "example" {
  name         = "My Salesforce Connection"
  organization = var.organization_id

  configuration = {
    domain   = "example.my.salesforce.com"
    password = var.salesforce_password
    username = var.salesforce_username
    token    = var.salesforce_token
  }
}

# Use the data source to discover the schema and its fields
data "polytomic_connection_schema" "accounts" {
  connection_id = polytomic_salesforce_connection.example.id
  schema_id     = "Account" # Salesforce object name
}

# Set custom primary keys for the schema
resource "polytomic_connection_schema_primary_keys" "accounts_pk" {
  connection_id = data.polytomic_connection_schema.accounts.connection_id
  schema_id     = data.polytomic_connection_schema.accounts.schema_id

  # Specify the field IDs that should be used as primary keys
  # You can find these IDs in the data source's fields attribute
  field_ids = [
    # Example: Use a custom field as primary key instead of the default Id field
    "CustomUniqueId__c",
  ]
}

# Example with composite primary key
resource "polytomic_connection_schema_primary_keys" "multi_field_pk" {
  connection_id = polytomic_salesforce_connection.example.id
  schema_id     = "CustomObject__c"

  # Use multiple fields as a composite primary key
  field_ids = [
    "CompanyId__c",
    "ProductId__c",
  ]
}
