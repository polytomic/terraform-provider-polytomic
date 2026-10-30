resource "polytomic_planetscale_vitess_connection" "planetscale_vitess" {
  name = "example"
  configuration = {
    hostname    = "aws.connect.psdb.cloud"
    keyspace    = "mykeyspace"
    ssh_host    = "bastion.example.com"
    tablet_type = "primary"
  }
}

