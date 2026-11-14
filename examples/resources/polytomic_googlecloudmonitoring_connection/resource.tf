resource "polytomic_googlecloudmonitoring_connection" "googlecloudmonitoring" {
  name = "example"
  configuration = {
    project_id = "my-project"
  }
}

