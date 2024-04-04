resource "polytomic_ironclad_connection" "ironclad" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

