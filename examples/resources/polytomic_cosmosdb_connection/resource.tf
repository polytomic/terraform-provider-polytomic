resource "polytomic_cosmosdb_connection" "cosmosdb" {
  name = "example"
  configuration = {
    uri = "https://my-account.documents.example.com:443"
    key = "cosmosdb-secret-key"
  }
}

