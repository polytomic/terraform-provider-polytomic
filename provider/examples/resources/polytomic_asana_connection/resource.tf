resource "polytomic_asana_connection" "asana" {
  name         = "example"
  configuration = {
    pat = "my-personal-access-token"
  }
}

