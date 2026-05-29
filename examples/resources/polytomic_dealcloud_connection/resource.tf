resource "polytomic_dealcloud_connection" "dealcloud" {
  name = "example"
  configuration = {
    host = "mycompany.dealcloud.com"
  }
}

