resource "polytomic_salesforce_connection" "salesforce" {
  name = "example"
  configuration = {
    client_id           = "a45gadsfdsaf47byor2ugfbhsgllpf12gf56gfds"
    client_secret       = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    connect_mode        = "api"
    domain              = "http://instance.my.salesforce.com"
    oauth_refresh_token = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
  }
}

