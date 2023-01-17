resource "polytomic_netsuite_connection" "netsuite" {
  name = "example"
  configuration = {
    account_id   = "my-account-id"
    consumer_key = "my-consumer-key"
  }
}

