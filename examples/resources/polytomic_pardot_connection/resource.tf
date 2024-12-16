resource "polytomic_pardot_connection" "pardot" {
  name = "example"
  configuration = {
    account_type     = "Production"
    business_unit_id = "1234567"
  }
}

