resource "polytomic_coupa_connection" "coupa" {
  name = "example"
  configuration = {
    host = "acme.coupahost.com"
  }
}

