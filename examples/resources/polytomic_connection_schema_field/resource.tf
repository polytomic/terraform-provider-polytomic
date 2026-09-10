# Example: Add and override fields on a connection schema
#
# User-defined fields are available on connections that support them, such as
# MongoDB, DynamoDB, Stripe, and file storage connections.

resource "polytomic_mongodb_connection" "orders" {
  name         = "Orders"
  organization = var.organization_id

  configuration = {
    hosts    = "mongo.example.com:27017"
    database = "shop"
    username = var.mongodb_username
    password = var.mongodb_password
  }
}

# Add a field extracted from a nested document at read time
resource "polytomic_connection_schema_field" "city" {
  connection_id = polytomic_mongodb_connection.orders.id
  schema_id     = "shop.orders"
  field_id      = "city"
  label         = "City"
  type          = "string"
  path          = "$.address.city"
}

# Override the type of a field the source already reports; the label is
# inherited from the source
resource "polytomic_connection_schema_field" "amount" {
  connection_id = polytomic_mongodb_connection.orders.id
  schema_id     = "shop.orders"
  field_id      = "amount"
  type          = "number"
}
