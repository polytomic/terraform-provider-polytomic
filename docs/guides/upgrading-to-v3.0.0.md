---
page_title: "Upgrading to v3.0.0"
subcategory: ""
description: |-
  Notes on upgrading the Polytomic provider to v3.0.0.
---

# Upgrading to v3.0.0

Version 3.0.0 moves the provider onto the Polytomic `2025-09-18` API version and refreshes every connection schema against it. Most of that work is internal, but two things can break an existing configuration:

1. The Sage Intacct connection was renamed from `polytomic_sageintacct_connection` to `polytomic_sage_intacct_connection`.
2. Seven connection types now require a configuration field that was previously optional or absent.

Terraform state written by v2.x is otherwise compatible. The provider's internal SDK types were renamed for the new API version, but every `tfsdk` tag was left alone, so no attribute names or state layouts changed as a result.

The provider now sends `X-Polytomic-Version: 2025-09-18` on every request, up from `2024-02-08`, and there is no setting to pin the older version.

## 1. Sage Intacct connection renamed

The Sage Intacct backend ID changed from `sageintacct` to `sage_intacct`, which renames both the resource and the data source:

| v2.x | v3.0.0 |
|---|---|
| `polytomic_sageintacct_connection` | `polytomic_sage_intacct_connection` |

The two schemas are otherwise identical, but Terraform refuses to move state between two different resource types:

```
Error: Invalid state move request

Cannot move polytomic_sageintacct_connection.example to
polytomic_sage_intacct_connection.example: resource types don't match.
```

So this needs a remove-and-import rather than a `terraform state mv`.

First, note the connection ID from the existing state:

```console
$ terraform state show polytomic_sageintacct_connection.example
# polytomic_sageintacct_connection.example:
resource "polytomic_sageintacct_connection" "example" {
    id   = "c7f3a1e2-...-9b4d"
    name = "Sage Intacct"
    ...
}
```

Rename the resource in your configuration:

```hcl
# Before (v2.x)
resource "polytomic_sageintacct_connection" "example" {
  name = "Sage Intacct"
  configuration = {
    company_id = var.sage_company_id
    # ...
  }
}

# After (v3.0.0)
resource "polytomic_sage_intacct_connection" "example" {
  name = "Sage Intacct"
  configuration = {
    company_id = var.sage_company_id
    # ...
  }
}
```

Then drop the old address from state and import the connection at the new one:

```console
$ terraform state rm polytomic_sageintacct_connection.example
$ terraform import polytomic_sage_intacct_connection.example c7f3a1e2-...-9b4d
```

`terraform state rm` only forgets the resource; it does not delete the connection in Polytomic.

Update any references to the old address — including `depends_on`, outputs, and `connection_id` arguments on models, syncs, and bulk syncs — to the new one. Because those references point at the same connection ID after the import, dependent resources should show no diff.

Finally, confirm the result:

```console
$ terraform plan
```

Sensitive fields are not returned by the API, so an imported connection may plan an in-place update that re-sends secrets from your configuration. That is expected.

## 2. Newly required configuration fields

Seven connection types now require a field. Four were previously optional, and three are new fields on existing connection types. In both cases the field selects how the connection authenticates or which account it targets, and the API now expects it explicitly.

| Connection | Field | v2.x | Accepted values | Previous default |
|---|---|---|---|---|
| `polytomic_apple_ads_connection` | `org_id` | not present | your Apple Ads organization ID | none |
| `polytomic_attio_connection` | `auth_method` | not present | `oauth`, `api_key` | none |
| `polytomic_googleworkspace_connection` | `auth_method` | optional | `oauth`, `service_account` | `oauth` |
| `polytomic_msads_connection` | `auth_method` | optional | `microsoft`, `google` | `microsoft` |
| `polytomic_msdynamics_connection` | `auth_method` | not present | `oauth`, `client_credentials` | `oauth` |
| `polytomic_salesforce_connection` | `connect_mode` | optional | `browser`, `clientcredentials`, `code`, `api` | `browser` |
| `polytomic_xero_connection` | `connect_mode` | optional | `browser`, `clientcredentials` | `browser` |

A configuration missing one of these fails at plan time:

```
Error: Incorrect attribute value type

  on main.tf line 10, in resource "polytomic_xero_connection" "example":
  10:   configuration = {
  11:     client_id     = "a"
  12:     client_secret = "b"
  13:   }

Inappropriate value for attribute "configuration": attribute "connect_mode"
is required.
```

Set the value that matches how the connection is already configured in Polytomic. Where the field previously had a default, an existing connection that never set it is using that default, so carrying the "Previous default" column above into your configuration preserves current behavior. `polytomic_apple_ads_connection.org_id` and `polytomic_attio_connection.auth_method` have no default — supply the value the connection actually uses.

```hcl
# Before (v2.x)
resource "polytomic_xero_connection" "example" {
  name = "Xero"
  configuration = {
    client_id     = var.xero_client_id
    client_secret = var.xero_client_secret
  }
}

# After (v3.0.0)
resource "polytomic_xero_connection" "example" {
  name = "Xero"
  configuration = {
    connect_mode  = "browser"
    client_id     = var.xero_client_id
    client_secret = var.xero_client_secret
  }
}
```

If you are unsure which value a given connection uses, check the connection in the Polytomic UI, or read it back after upgrading with `terraform state show`.

`polytomic_msdynamics_connection` gained `client_credentials_client_id`, `client_credentials_client_secret`, and `tenant_id` alongside its new `auth_method`. Set those three only when `auth_method = "client_credentials"`.

## Relaxed requirements

These fields became optional. No configuration changes are needed, but you may now omit them:

| Connection | Field(s) |
|---|---|
| `polytomic_dbtprojectrepository_connection` | `commit_exposures`, `oauth_access_token`, `repository` |
| `polytomic_dropbox_connection` | `bucket` |
| `polytomic_github_connection` | `oauth_access_token` |
| `polytomic_salesforce_connection` | `client_id`, `client_secret` |

## Importer

The `polytomic` importer generates configuration from your existing Polytomic resources, so regenerating with the v3.0.0 importer produces the renamed Sage Intacct resource and fills in the newly required fields from the live connection. If you maintain generated configuration, regenerating is the quickest path through this upgrade — you will still need to re-import Sage Intacct state as described above.

## New connection types

v3.0.0 adds resources and data sources for: Amazon Ads, Amazon RDS PostgreSQL, Anrok, Bill, Brex, Clarify, Claude Analytics, Coupa, Fakturownia, Granola, InvoiceOcean, LinkedIn Company Pages, Measure, Meridian, Neon, Nooks, PlanetScale Vitess, Podscribe, Railway, Resend, Rokt, S3-Compatible, Snapchat Ads, StackAdapt, Supabase, Vibe, X Ads, and Zoom.

Apple Ads Attribution and Polytomic Harbor are available as data sources only; they have no user-settable configuration.
