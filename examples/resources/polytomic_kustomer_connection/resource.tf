resource "polytomic_kustomer_connection" "kustomer" {
  name = "example"
  configuration = {
    domain = "my-domain.example.com"
  }
}

