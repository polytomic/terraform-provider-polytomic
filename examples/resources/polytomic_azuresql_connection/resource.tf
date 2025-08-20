resource "polytomic_azuresql_connection" "azuresql" {
  name = "example"
  configuration = {
    access_key     = "abcdefghijklmnopqrstuvwxyz0123456789/+ABCDEabcdefghijklmnopqrstuvwxyz0123456789/+ABCDE=="
    account_name   = "account"
    container_name = "container"
    database       = "sampledb"
    hostname       = "example.database.windows.net"
    password       = "secret"
    ssh_host       = "bastion.example.com"
    username       = "user"
  }
}

