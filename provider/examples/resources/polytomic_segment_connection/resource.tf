resource "polytomic_segment_connection" "segment" {
  name         = "example"
  configuration = {
    write_key = "my-write-key"
  }
}

