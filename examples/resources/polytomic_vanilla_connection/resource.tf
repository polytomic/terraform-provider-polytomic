resource "polytomic_vanilla_connection" "vanilla" {
  name = "example"
  configuration = {
    api_key = "my-api-key"
    domain  = "example.com"
  }
}

