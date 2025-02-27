---
# generated by https://github.com/fbreckle/terraform-plugin-docs
page_title: "polytomic_gcs_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Google Cloud Storage Connection
---

# polytomic_gcs_connection (Resource)

Google Cloud Storage Connection

## Example Usage

```terraform
resource "polytomic_gcs_connection" "gcs" {
  name = "example"
  configuration = {
    bucket            = "my-bucket"
    single_table_name = "collection"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `configuration` (Attributes) (see [below for nested schema](#nestedatt--configuration))
- `name` (String)

### Optional

- `force_destroy` (Boolean) Indicates whether dependent models, syncs, and bulk syncs should be cascade
deleted when this connection is destroy.

  This only deletes other resources when the connection is destroyed, not when
setting this parameter to `true`. Once this parameter is set to `true`, there
must be a successful `terraform apply` run before a destroy is required to
update this value in the resource state. Without a successful `terraform apply`
after this parameter is set, this flag will have no effect. If setting this
field in the same operation that would require replacing the connection or
destroying the connection, this flag will not work. Additionally when importing
a connection, a successful `terraform apply` is required to set this value in
state before it will take effect on a destroy operation.
- `organization` (String) Organization ID

### Read-Only

- `id` (String) Google Cloud Storage Connection identifier

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

Required:

- `bucket` (String)
- `service_account` (String, Sensitive) Service Account Key

Optional:

- `client_email` (String) Service account identity
- `is_single_table` (Boolean) Files are time-based snapshots

    Treat the files as a single table.
- `project_id` (String) Service account project ID
- `single_table_file_format` (String) File format
- `single_table_name` (String) Collection name
- `skip_lines` (Number) Skip first lines

    Skip first N lines of each CSV file.


