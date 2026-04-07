# Upgrading to v1.4.0

Version 1.4.0 redesigns the `filters`, `overrides`, and `override_fields` attributes on the `polytomic_sync` resource. These changes make filters and overrides easier to use by replacing opaque field UUIDs with human-readable source references.

## Migration steps

After upgrading the provider, run `terraform plan` to see what changes are needed. Terraform will report errors for any syncs using the old attribute shapes. Update your configuration as described below, then run `terraform apply`.

Because the attribute shapes changed, Terraform will show filters and overrides as being replaced (removed + added). The underlying sync configuration on the server is unchanged.

## `filters`: Use source references instead of field UUIDs

The `field_id` and `field_type` attributes have been replaced with a `source` block that references the model and field by name. The server resolves the field UUID automatically.

```hcl
# Before (v1.3.x)
filters = [{
  field_id   = "a1b2c3d4-..."   # opaque field UUID
  field_type = "Model"
  function   = "Equality"
  value      = jsonencode("test@example.com")
}]

# After (v1.4.0)
filters = [{
  source = {
    model_id = polytomic_model.example.id
    field    = "email"
  }
  function = "Equality"
  value    = jsonencode("test@example.com")
}]
```

## `target_filters`: Target-type filters moved to a separate attribute

Filters that previously used `field_type = "Target"` are now configured with the new `target_filters` attribute. Target filters only work with syncs in `update` mode. Use `target.filter_logic` (not `filter_logic`) to combine target filters.

```hcl
# Before (v1.3.x)
filters = [{
  field_id   = "status"
  field_type = "Target"
  function   = "Equality"
  value      = jsonencode("active")
}]

# After (v1.4.0)
target_filters = [{
  field    = "status"
  function = "Equality"
  value    = jsonencode("active")
}]
```

## `overrides`: Use source references instead of field UUIDs

Same change as filters: `field_id` is replaced with a `source` block.

```hcl
# Before (v1.3.x)
overrides = [{
  field_id = "a1b2c3d4-..."
  function = "IsNull"
  override = jsonencode("fallback@example.com")
}]

# After (v1.4.0)
overrides = [{
  source = {
    model_id = polytomic_model.example.id
    field    = "email"
  }
  function = "IsNull"
  override = jsonencode("fallback@example.com")
}]
```

## `override_fields`: Remove the `source` block

The `source` block on `override_fields` was never used by the server and has been removed. Delete the `source` block from any `override_fields` entries.

```hcl
# Before (v1.3.x)
override_fields = [{
  source = {
    model_id = polytomic_model.example.id
    field    = "email"
  }
  target         = "some_field"
  override_value = "static_value"
}]

# After (v1.4.0)
override_fields = [{
  target         = "some_field"
  override_value = "static_value"
}]
```

## `target.search_values`: Removed

The `search_values` attribute on the `target` block has been removed. It was internal application state that the server derives from the `object` field. Remove it from your configuration; no replacement is needed.

```hcl
# Before (v1.3.x)
target = {
  connection_id = polytomic_postgresql_connection.example.id
  object        = "public.users"
  search_values = jsonencode({
    "schema" = "public"
    "table"  = "users"
  })
}

# After (v1.4.0)
target = {
  connection_id = polytomic_postgresql_connection.example.id
  object        = "public.users"
}
```

If you use the importer to generate Terraform configuration, it will no longer emit `search_values`.

## `filter_logic`: Labels use letters

Filter logic expressions reference filters by letter labels (A, B, C, ...) assigned by the server. When using `filter_logic`, provide explicit labels on your filters:

```hcl
filters = [
  {
    source   = { model_id = polytomic_model.example.id, field = "email" }
    function = "IsNotNull"
    label    = "A"
  },
  {
    source   = { model_id = polytomic_model.example.id, field = "status" }
    function = "Equality"
    value    = jsonencode("active")
    label    = "B"
  }
]

filter_logic = "A AND B"
```

## Recovering from state errors

If `terraform plan` fails with state errors after upgrading, you can re-import the affected sync:

```bash
# Remove the sync from state (does NOT delete from Polytomic)
terraform state rm polytomic_sync.example

# Re-import with the sync ID
terraform import polytomic_sync.example <sync-id>
```

The sync ID can be found in the Polytomic UI or in the previous state file.
