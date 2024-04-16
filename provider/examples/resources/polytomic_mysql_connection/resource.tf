resource "polytomic_mysql_connection" "mysql" {
  name         = "example"
  configuration = {
    hostname = "mysql.example.com"
    account = "acme"
    passwd = "super-secret-password"
    dbname = "db"
    port = 3306
    change_detection = false
  }
}

