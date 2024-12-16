resource "polytomic_recharge_connection" "recharge" {
  name = "example"
  configuration = {
    api_key = "secret"
  }
}

