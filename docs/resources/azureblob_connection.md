---
page_title: "polytomic_azureblob_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Azure Blob Storage Connection
---

# polytomic_azureblob_connection (Resource)

Azure Blob Storage Connection

For detailed configuration guidance, see the [Azure Blob Storage connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/azureblob).

## Example Usage

```terraform
resource "polytomic_azureblob_connection" "azureblob" {
  name = "example"
  configuration = {
    access_key               = "abcdefghijklmnopqrstuvwxyz0123456789/+ABCDEabcdefghijklmnopqrstuvwxyz0123456789/+ABCDE=="
    account_name             = "account"
    container_name           = "container"
    oauth_refresh_token      = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    single_table_file_format = "csv"
    single_table_name        = "collection"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) Azure Blob Storage Connection identifier.
- `force_destroy` (Boolean, Optional) Indicates whether dependent models, syncs, and bulk syncs should be
cascade-deleted when this connection is destroyed.

    This only deletes other resources when the connection is destroyed, not when
setting this parameter to `true`. Once this parameter is set to `true`, there
must be a successful `terraform apply` run before a destroy is required to
update this value in the resource state. Without a successful `terraform apply`
after this parameter is set, this flag will have no effect. If setting this
field in the same operation that would require replacing the connection or
destroying the connection, this flag will not work. Additionally when importing
a connection, a successful `terraform apply` is required to set this value in
state before it will take effect on a destroy operation.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

- `access_key` (String, Sensitive, Optional) Access Key
- `account_name` (String, Required) Account Name
- `auth_method` (String, Required) Authentication method Valid values: <code>access_key</code> (Access Key), <code>client_credentials</code> (Client Credentials), <code>oauth</code> (Oauth). Default: <code>access_key</code>.
- `client_id` (String, Optional) Client ID
- `client_secret` (String, Sensitive, Optional) Client Secret
- `container_name` (String, Required) Container Name
- `csv_has_headers` (Boolean, Optional) CSV files have headers

    Whether CSV files have a header row with field names. Default: <code>true</code>.
- `directory_glob_pattern` (String, Optional) Tables glob path
- `is_directory_snapshot` (Boolean, Optional) Multi-directory multi-table Default: <code>false</code>.
- `is_single_table` (Boolean, Optional) Files are time-based snapshots

    Treat the files as a single table. Default: <code>false</code>.
- `oauth_refresh_token` (String, Sensitive, Optional)
- `single_table_file_format` (String, Optional) File format Valid values: <code>csv</code> (CSV), <code>json</code> (JSON), <code>parquet</code> (Parquet). Default: <code>csv</code>.
- `single_table_file_formats` (Set of String, Optional) File formats

    File formats that may be present across different tables Default: <code>[[csv]]</code>.
- `single_table_name` (String, Optional) Collection name
- `skip_lines` (Number, Optional) Skip first lines

    Skip first N lines of each CSV file. Default: <code>0</code>.
- `tenant_id` (String, Optional) Tenant ID

