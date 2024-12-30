data "polytomic_bulk_source" "source" {
  connection_id = "aab123aa-27f3-abc1-9999-abcde123a4aa"
}

data "polytomic_bulk_destination" "dest" {
  connection_id = "bbd321bb-abc1-27f3-1111-abcde123a1bb"
}

resource "polytomic_bulk_sync" "sync" {
  name   = "Terraform Bulk Sync"
  active = true
  mode   = "replicate"

  schedule = {
    frequency = "manual"
  }
  source = {
    connection_id = data.polytomic_bulk_source.source.connection_id
  }
  destination = {
    connection_id = data.polytomic_bulk_destination.dest.connection_id
    configuration = {
      "schema" = "terraform"
    }
  }
  schemas = data.polytomic_bulk_source.source.schemas.*.id

  discovery = {
    enabled = false
  }
}
