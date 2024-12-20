resource "polytomic_snowflake_connection" "snowflake" {
  name = "example"
  configuration = {
    account   = "uc193736182"
    dbname    = "database_name"
    password  = "password"
    username  = "user"
    warehouse = "warehouse"
  }
}

