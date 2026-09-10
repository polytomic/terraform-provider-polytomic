---
page_title: "polytomic_msdynamics_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Microsoft Dynamics 365 Connection
---

# polytomic_msdynamics_connection (Resource)

Microsoft Dynamics 365 Connection

For detailed configuration guidance, see the [Microsoft Dynamics 365 connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/msdynamics).

## Example Usage

```terraform
resource "polytomic_msdynamics_connection" "msdynamics" {
  name = "example"
  configuration = {
    client_credentials_client_id     = "a45gadsfdsafbyorxhugfbhsgllpf12gf56gfds"
    client_credentials_client_secret = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    client_id                        = "a45gadsfdsafbyorxhugfbhsgllpf12gf56gfds"
    client_secret                    = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    oauth_refresh_token              = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    tenant_id                        = "3e03e565-ca33-4ef5-8e19-db300c655a40"
  }
}
```

## Schema

### Required

- `name` (String)
- `configuration` (Attributes) See [below for nested schema](#nestedatt--configuration).

### Optional

- `organization` (String) Organization ID.
- `force_destroy` (Boolean) Indicates whether dependent models, syncs, and bulk syncs should be
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

### Read-Only

- `id` (String) Microsoft Dynamics 365 Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `auth_method` (String) Authentication method Valid values: <code>oauth</code> (OAuth), <code>client_credentials</code> (Client credentials). Default: <code>oauth</code>.
- `dynamics_url` (String) Dynamics URL

#### Optional

- `client_credentials_client_id` (String, Sensitive) Client ID
- `client_credentials_client_secret` (String, Sensitive) Client secret
- `client_id` (String, Sensitive) Custom OAuth application client ID for the delegated flow.
- `client_secret` (String, Sensitive) Custom OAuth application client secret for the delegated flow.
- `oauth_refresh_token` (String, Sensitive)
- `tenant_id` (String) Directory (tenant) ID


