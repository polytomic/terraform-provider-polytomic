resource "polytomic_paycor_connection" "paycor" {
  name = "example"
  configuration = {
    client_id        = "your_client_id"
    client_secret    = "your_client_secret"
    scopes           = "profile"
    subscription_key = "your_subscription_key"
  }
}

