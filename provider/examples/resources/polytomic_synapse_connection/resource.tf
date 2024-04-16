resource "polytomic_synapse_connection" "synapse" {
  name         = "example"
  configuration = {
    hostname = "host.example.com"
    username = "user"
    password = "password"
    database = "database"
    port = 5439
  }
}

