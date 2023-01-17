resource "polytomic_mysql_connection" "mysql" {
  name = "example"
  configuration = {
    hostname = "mysql.example.com"
    account  = "acme"
    dbname   = "db"
    port     = 3306
  }
}

