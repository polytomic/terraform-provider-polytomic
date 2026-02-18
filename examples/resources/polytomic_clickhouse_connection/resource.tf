resource "polytomic_clickhouse_connection" "clickhouse" {
  name = "example"
  configuration = {
    database = "default"
    hostname = "clickhouse.example.com"
    ssh_host = "bastion.example.com"
    username = "default"
  }
}

