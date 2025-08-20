resource "polytomic_snowflake_connection" "snowflake" {
  name = "example"
  configuration = {
    account   = "FRXJLEC-UJA94780"
    dbname    = "database_name"
    password  = "password"
    username  = "user"
    warehouse = "warehouse"
  }
}

