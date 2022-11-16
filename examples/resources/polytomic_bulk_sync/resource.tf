data "polytomic_bulk_source" "source" {
  connection_id = "aab123aa-27f3-abc1-9999-abcde123a4aa"
}

data "polytomic_bulk_destination" "dest" {
  connection_id = "bbd321bb-abc1-27f3-1111-abcde123a1bb"
}

resource "polytomic_bulk_sync" "blah" {
  name                 = "Terraform Bulk Sync"
  source_connection_id = data.polytomic_bulk_source.source.connection_id
  dest_connection_id   = data.polytomic_bulk_destination.dest.connection_id
  active               = true
  discover             = false
  mode                 = "replicate"
  schedule = {
    frequency = "manual"
  }
  schemas = [
    for s in data.polytomic_bulk_source.source.schemas : {
      name    = s.name
      id      = s.id
      enabled = true
    }
  ]
  dest_configuration = {
    "dataset" = "terraform"
  }
}
