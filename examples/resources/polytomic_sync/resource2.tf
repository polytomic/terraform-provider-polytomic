resource "polytomic_sync" "sync2" {
  name = "Terraform BigQuery my_new_table sync"
  mode = "updateOrCreate"
  schedule = {
    frequency = "manual"
  }
  fields = [
    {
      source = {
        field    = "last_name"
        model_id = polytomic_model.model.id
      }
      newField = true
      target   = "last_name"
    },

  ]
  target = {
    connection_id = "19962644-780b-11ed-aeca-ea7534cffcab"
    object        = "__pt_new_target"
    search_values = {
      "dataset" = "awesome"
      "table"   = "new_table"
    }
    new_name = "awesome.table"
  }
  identity = {
    source = {
      field    = "first_name"
      model_id = polytomic_model.model.id
    }
    target = "first_name"
  }

}

