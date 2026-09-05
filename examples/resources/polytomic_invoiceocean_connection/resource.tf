resource "polytomic_invoiceocean_connection" "invoiceocean" {
  name = "example"
  configuration = {
    subdomain = "acme"
  }
}

