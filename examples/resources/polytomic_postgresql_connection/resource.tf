resource "polytomic_postgresql_connection" "postgresql" {
  name = "example"
  configuration = {
    database    = "sampledb"
    hostname    = "database.example.com"
    password    = "password"
    publication = "polytomic"
    ssh_host    = "bastion.example.com"
    username    = "postgres"
  }
}

