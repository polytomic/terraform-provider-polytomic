resource "polytomic_synapse_connection" "synapse" {
  name = "example"
  configuration = {
    database = "yourdatabase"
    hostname = "yourserver.sql.azuresynapse.net"
    password = "secret"
    username = "user"
  }
}

