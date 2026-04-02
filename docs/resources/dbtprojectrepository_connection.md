---
page_title: "polytomic_dbtprojectrepository_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  dbt Project Repository Connection
---

# polytomic_dbtprojectrepository_connection (Resource)

dbt Project Repository Connection

For detailed configuration guidance, see the [dbt Project Repository connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/dbtprojectrepository).

## Example Usage

```terraform
resource "polytomic_dbtprojectrepository_connection" "dbtprojectrepository" {
  name = "example"
  configuration = {
    oauth_access_token  = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    oauth_refresh_token = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
  }
}
```

## Schema

- `name` (String, Required)
- `configuration` (Attributes, Required) See [below for nested schema](#nestedatt--configuration).
- `organization` (String, Optional) Organization ID.
- `id` (String, Read-only) dbt Project Repository Connection identifier.
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

- `branch` (String, Optional) Exposures branch
- `commit_exposures` (Boolean, Required) Commit exposures file
- `connected_user` (String, Optional) Connected user
- `latest_commit` (Attributes, Optional) Most recent exposures commit See [below for nested schema](#nestedatt--configuration--latest_commit).
- `oauth_access_token` (String, Sensitive, Required)
- `oauth_refresh_token` (String, Sensitive, Optional)
- `oauth_token_expiry` (String, Optional)
- `repository` (String, Required) dbt project repository

    Only repositories with the Polytomic app installed will be listed.

<a id="nestedatt--configuration--latest_commit"></a>
### Nested Schema for `configuration.latest_commit`

- `href` (String, Optional)
- `text` (String, Optional)

