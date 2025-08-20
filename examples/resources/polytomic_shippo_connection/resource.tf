resource "polytomic_shippo_connection" "shippo" {
  name = "example"
  configuration = {
    api_key = "token"
  }
}

