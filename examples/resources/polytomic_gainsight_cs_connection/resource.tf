resource "polytomic_gainsight_cs_connection" "gainsight_cs" {
  name = "example"
  configuration = {
    domain = "company.gainsightcloud.com"
  }
}

