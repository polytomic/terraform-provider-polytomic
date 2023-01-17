resource "polytomic_snowflake_connection" "snowflake" {
  name = "example"
  configuration = {
    account   = "acme"
    username  = "user"
    password  = "secret"
    dbname    = "db"
    warehouse = "warehouse"
  }
}

