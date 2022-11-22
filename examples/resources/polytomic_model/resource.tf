resource "polytomic_model" "model" {
  name          = "Terraform model"
  connection_id = "bbd321bb-abc1-27f3-1111-abcde123a1bb"

  configuration = {
    "database"   = "acme"
    "collection" = "users"
  }

}
