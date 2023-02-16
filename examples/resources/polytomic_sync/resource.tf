resource "polytomic_sync" "sync" {
  name = "Terraform sync"
  mode = "replace"
  schedule = {
    frequency = "manual"
  }
  fields = [
    {
      source = {
        field    = "email"
        model_id = "516a4fbe-464a-4678-8153-ec53d2e4bdd5"
      }
      target = "record"
    }
  ]
  target = {
    connection_id = "bbd321bb-abc1-27f3-1111-abcde123a1bb"
    object        = "test"
    configuration = jsonencode({
      "format" = "csv"
    })
  }

}
