resource "polytomic_azureblob_connection" "azureblob" {
  name = "example"
  configuration = {
    access_key     = "abcdefghijklmnopqrstuvwxyz0123456789/+ABCDEabcdefghijklmnopqrstuvwxyz0123456789/+ABCDE=="
    account_name   = "account"
    container_name = "container"
  }
}

