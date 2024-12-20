resource "polytomic_dbtcloud_connection" "dbtcloud" {
  name = "example"
  configuration = {
    token = "secret"
    url   = "https://cloud.getdbt.com"
  }
}

