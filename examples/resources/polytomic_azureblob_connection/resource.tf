resource "polytomic_azureblob_connection" "azureblob" {
  name = "example"
  configuration = {
    account_name   = "my-account"
    container_name = "my-container"
  }
}

