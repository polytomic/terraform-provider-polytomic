resource "polytomic_bigquery_connection" "bigquery" {
  organization = polytomic_organization.acme.id
  name         = "example"
  configuration = {
    service_account_credentials = data.account_credentials.json
    location                    = "us-central1"
  }
}

