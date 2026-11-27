resource "polytomic_sap_business_one_for_hana_connection" "sap_business_one_for_hana" {
  name = "example"
  configuration = {
    hostname = "hana.example.com"
  }
}

