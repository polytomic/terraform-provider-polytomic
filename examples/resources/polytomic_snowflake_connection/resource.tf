resource "polytomic_snowflake_connection" "snowflake" {
  organization = polytomic_organization.acme.id
  name         = "example"
  configuration = {
  }
}

