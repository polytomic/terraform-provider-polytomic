resource "polytomic_neon_connection" "neon" {
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

