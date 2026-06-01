---
page_title: "Upgrading to v2.0.0"
subcategory: ""
description: |-
  Notes on upgrading the Polytomic provider to v2.0.0.
---

# Upgrading to v2.0.0

Version 2.0.0 removes server-managed internal fields from connection configurations. These fields were never meant to be set or read through Terraform — they track internal state such as whether an OAuth flow has completed (`authenticated`) or when an OAuth token expires (`oauth_token_expiry`). Previously they were generated as `Optional` + `Computed`, so they appeared in the schema and documentation and were even assignable, despite having no effect.

Unlike the server-managed fields addressed in [v1.5.0](upgrading-to-v1.5.0), which carry useful information (such as the connected user) and were made read-only, these fields carry no configuration or reference value. They are now removed entirely.

This affects 28 fields across 27 connection types. Fields that are hidden in the Polytomic UI but settable through the API — such as `client_id`, `client_secret`, and `oauth_access_token` — are unchanged.

## Migration steps

After upgrading, run `terraform plan`.

If a configuration assigns one of the removed fields, Terraform returns an error like:

```
Error: Unsupported argument
  on main.tf line 5, in resource "polytomic_github_connection" "example":
   5:     authenticated = true

An argument named "authenticated" is not expected here.
```

Remove the assignment from your configuration. These fields were managed by Polytomic, so removing them has no effect on the connection.

```hcl
# Before (v1.5.x)
resource "polytomic_github_connection" "example" {
  name = "GitHub"
  configuration = {
    oauth_access_token = var.github_token
    authenticated      = true # remove this line
  }
}

# After (v2.0.0)
resource "polytomic_github_connection" "example" {
  name = "GitHub"
  configuration = {
    oauth_access_token = var.github_token
  }
}
```

Terraform drops the removed fields from existing state automatically on the next refresh; no manual state changes are required.

If another resource or output references a removed field (for example `polytomic_github_connection.example.configuration.authenticated`), remove that reference — the value is no longer exported.

## Importer

If you generated your configuration with the `polytomic` importer, regenerate it with the v2.0.0 importer to drop the removed fields automatically.

## Affected fields

| Connection | Field(s) |
|---|---|
| polytomic_dbtprojectrepository_connection | oauth_token_expiry |
| polytomic_dialpad_connection | oauth_token_expiry |
| polytomic_dropbox_connection | oauth_token_expiry |
| polytomic_github_connection | authenticated |
| polytomic_gmail_connection | oauth_token_expiry |
| polytomic_gong_connection | oauth_token_expiry |
| polytomic_googleads_connection | oauth_token_expiry |
| polytomic_googleanalytics_connection | oauth_token_expiry |
| polytomic_googlesearchconsole_connection | oauth_token_expiry |
| polytomic_googleslides_connection | oauth_token_expiry |
| polytomic_googleworkspace_connection | oauth_token_expiry |
| polytomic_gsheets_connection | oauth_token_expiry |
| polytomic_linkedinads_connection | oauth_token_expiry |
| polytomic_marketo_connection | oauth_token_expiry |
| polytomic_msads_connection | oauth_token_expiry |
| polytomic_msdynamics_connection | oauth_token_expiry |
| polytomic_outreach_connection | oauth_token_expiry |
| polytomic_quickbooks_connection | oauth_token_expiry |
| polytomic_ramp_connection | oauth_token_expiry |
| polytomic_redditads_connection | oauth_token_expiry |
| polytomic_sageintacct_connection | oauth_token_expiry |
| polytomic_salesloft_connection | oauth_token_expiry |
| polytomic_typeform_connection | oauth_token_expiry |
| polytomic_upfluence_connection | oauth_refresh_token, oauth_token_expiry |
| polytomic_youtubeanalytics_connection | oauth_token_expiry |
| polytomic_zendesk_support_connection | oauth_token_expiry |
| polytomic_zoho_crm_connection | oauth_token_expiry |
