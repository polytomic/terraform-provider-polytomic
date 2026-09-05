resource "polytomic_planetscale_vitess_connection" "planetscale_vitess" {
  name = "example"
  configuration = {
    database = "mydb"
    hostname = "aws.connect.psdb.cloud"
    ssh_host = "bastion.example.com"
  }
}

