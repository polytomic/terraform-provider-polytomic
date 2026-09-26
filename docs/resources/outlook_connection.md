---
page_title: "polytomic_outlook_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Microsoft Outlook 365 Connection
---

# polytomic_outlook_connection (Resource)

Microsoft Outlook 365 Connection

For detailed configuration guidance, see the [Microsoft Outlook 365 connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/outlook).

## Example Usage

```terraform
resource "polytomic_outlook_connection" "outlook" {
  name = "example"
  configuration = {
    client_credentials_client_id     = "eb669428-1854-4cb1-a560-403e05b8acbf"
    client_credentials_client_secret = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
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

- `id` (String) Microsoft Outlook 365 Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `auth_method` (String) Authentication method Valid values: <code>oauth</code> (OAuth), <code>client_credentials</code> (Client credentials). Default: <code>oauth</code>.

#### Optional

- `additional_mailboxes` (String) Shared mailboxes

    Comma-separated shared mailbox addresses to read alongside your mailbox
- `auto_add_mailboxes` (Boolean) Automatically add new mailboxes
- `client_credentials_client_id` (String) Client ID
- `client_credentials_client_secret` (String, Sensitive) Client secret
- `client_id` (String, Sensitive)
- `client_secret` (String, Sensitive)
- `mailboxes` (Attributes Set) See [below for nested schema](#nestedatt--configuration--mailboxes).
- `oauth_refresh_token` (String, Sensitive)
- `tenant_id` (String) Directory (tenant) ID

#### Read-Only

- `user_email` (String) Connected user's email


<a id="nestedatt--configuration--mailboxes"></a>
### Nested Schema for `configuration.mailboxes`

#### Optional

- `label` (String)
- `value` (String)


