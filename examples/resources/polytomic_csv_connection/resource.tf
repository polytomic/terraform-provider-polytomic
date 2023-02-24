resource "polytomic_csv_connection" "example" {
  name = "Example"
  configuration = {
    auth = {
      "oauth" : {
        "client_id" : "client_id",
        "client_secret" : "client_secret",
        "extra_form_data" : [],
        "token_endpoint" : "https://example.com/oauth/token"
      }
    }
    headers = [
      {
        "name"  = "foo"
        "value" = "bar"
      }
    ]
    parameters = [
      {
        "name"  = "foo"
        "value" = "bar"
      }
    ]
    url = "https://example.com/csv"
  }
}
