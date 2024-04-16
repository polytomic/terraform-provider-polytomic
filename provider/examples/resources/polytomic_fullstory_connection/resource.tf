resource "polytomic_fullstory_connection" "fullstory" {
  name         = "example"
  configuration = {
    api_key = "my-api-key"
  }
}

