resource "polytomic_sharepoint_excel_connection" "sharepoint_excel" {
  name = "example"
  configuration = {
    client_id           = "eb669428-1854-4cb1-a560-403e05b8acbf"
    client_secret       = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    oauth_access_token  = "{access token}"
    oauth_refresh_token = "{refresh token}"
    oauth_token_expiry  = "2023-11-21T21:48:51Z"
  }
}

