resource "polytomic_azureblob_connection" "azureblob" {
  name = "example"
  configuration = {
    access_key               = "abcdefghijklmnopqrstuvwxyz0123456789/+ABCDEabcdefghijklmnopqrstuvwxyz0123456789/+ABCDE=="
    account_name             = "account"
    container_name           = "container"
    oauth_refresh_token      = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    single_table_file_format = "csv"
    single_table_name        = "collection"
  }
}

