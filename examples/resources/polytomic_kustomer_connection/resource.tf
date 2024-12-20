resource "polytomic_kustomer_connection" "kustomer" {
  name = "example"
  configuration = {
    apikey = "secret"
    domain = "polytomic"
  }
}

