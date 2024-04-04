resource "polytomic_kustomer_connection" "kustomer" {
  name         = "example"
  configuration = {
    apikey = "my-api-key"
    domain = "my-domain.example.com"
  }
}

