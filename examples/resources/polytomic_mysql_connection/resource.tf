resource "polytomic_mysql_connection" "mysql" {
  name = "example"
  configuration = {
    account  = "admin"
    dbname   = "mydb"
    hostname = "database.example.com"
    passwd   = "password"
    ssh_host = "bastion.example.com"
  }
}

