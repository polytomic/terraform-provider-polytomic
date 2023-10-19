resource "polytomic_dittofeed_connection" "dittofeed" {
  name = "example"
  configuration = {
    url       = "https://example.dittofeed.com"
    write_key = "my-write-key"
  }
}

