resource "polytomic_sage_intacct_connection" "sage_intacct" {
  name = "example"
  configuration = {
    application_id      = "c873a91dfd9183d78143.app.sage.com"
    client_secret       = "f4d4b3b9010a33a83928dd5035c541d7cc1be91b"
    oauth_refresh_token = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
  }
}

