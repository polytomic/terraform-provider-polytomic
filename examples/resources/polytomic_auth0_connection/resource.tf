resource "polytomic_auth0_connection" "auth0" {
  name = "example"
  configuration = {
    client_id     = "jI2Zem1Yzxy8s8s..."
    client_secret = "cB6NNPhR12R8pJ7M..."
    domain        = "dev-g1ce1rt9.us.auth0.com"
  }
}

