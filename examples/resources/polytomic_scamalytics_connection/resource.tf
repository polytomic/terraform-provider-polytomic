resource "polytomic_scamalytics_connection" "scamalytics" {
  name = "example"
  configuration = {
    endpoint = "https://api9.scamalytics.com/xyz"
  }
}

