resource "polytomic_planetscale_postgres_connection" "planetscale_postgres" {
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

