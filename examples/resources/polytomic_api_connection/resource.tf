resource "polytomic_api_connection" "example" {
  name = "Example"
  configuration = {
    auth = {
      "header" : {
        "name" : "foo",
        "value" : "bar"
      },
    }
    url = "https://example.com"
  }
}
