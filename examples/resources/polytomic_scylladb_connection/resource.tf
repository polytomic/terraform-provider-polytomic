resource "polytomic_scylladb_connection" "scylladb" {
  name = "example"
  configuration = {
    hosts    = "scylla.example.com"
    password = "password"
    ssh_host = "bastion.example.com"
    username = "scylla"
  }
}

