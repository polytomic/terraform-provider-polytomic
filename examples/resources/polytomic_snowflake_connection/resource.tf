resource "polytomic_snowflake_connection" "snowflake" {
  name = "example"
  configuration = {
    account   = "acme"
    username  = "user"
    dbname    = "db"
    warehouse = "warehouse"
  }
}

