resource "polytomic_bigquery_connection" "bigquery" {
  name = "example"
  configuration = {
    service_account_credentials = data.account_credentials.json
    location                    = "us-central1"
  }
}

