resource "polytomic_vanilla_connection" "vanilla" {
  name = "example"
  configuration = {
    api_key = "secret"
    domain  = "yourcompany.vanillacommunities.com"
  }
}

