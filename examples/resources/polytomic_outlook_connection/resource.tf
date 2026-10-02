resource "polytomic_outlook_connection" "outlook" {
  name = "example"
  configuration = {
    client_credentials_client_id     = "eb669428-1854-4cb1-a560-403e05b8acbf"
    client_credentials_client_secret = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    tenant_id                        = "3e03e565-ca33-4ef5-8e19-db300c655a40"
  }
}

