---
page_title: "polytomic_googlecloudmonitoring_connection Resource - terraform-provider-polytomic"
subcategory: "Connections"
description: |-
  Google Cloud Monitoring Connection
---

# polytomic_googlecloudmonitoring_connection (Resource)

Google Cloud Monitoring Connection

For detailed configuration guidance, see the [Google Cloud Monitoring connection guide](https://apidocs.polytomic.com/guides/configuring-your-connections/connections/googlecloudmonitoring).

## Example Usage

```terraform
resource "polytomic_googlecloudmonitoring_connection" "googlecloudmonitoring" {
  name = "example"
  configuration = {
    project_id = "my-project"
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

- `id` (String) Google Cloud Monitoring Connection identifier.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

#### Required

- `service_account` (String, Sensitive) Service account key

    JSON key of a service account that can read Cloud Monitoring metrics.

#### Optional

- `custom_reports` (String) Custom reports

    Custom reports, one per line, each written as a report name:metric type|resource type:column list.
- `project_id` (String) Project ID

    ID of the project whose metrics are read. Defaults to the service account's project.

#### Read-Only

- `client_email` (String) Service account identity

    Email address of the service account.


