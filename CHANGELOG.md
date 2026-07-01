## v2.0.0 (1 July 2026)

BREAKING CHANGES:

- Removed server-managed internal fields from connection configurations. Fields such as `authenticated` (GitHub) and `oauth_token_expiry` were generated as `Optional` + `Computed` but tracked internal OAuth state and had no effect when set. This removes 28 fields across 27 connection types. Fields that are hidden in the UI but settable through the API (`client_id`, `client_secret`, `oauth_access_token`) are unchanged. See the [v2.0.0 upgrade guide](docs/guides/upgrading-to-v2.0.0.md) for the full list and migration steps.

- Added support for new connection types:
  - Amazon Keyspaces
  - Apple Ads
  - Baseten
  - CloudTalk
  - Google Search Ads 360
  - OpenAI Ads
  - Rillet
  - Wavelength
- Updated connection schemas for Asana, CSV URL, Google Ads, HTTP API, HTTP Enrichment, Knock, Plain, ScyllaDB, Tixr, Webhook, and Zendesk Support.

## v1.5.5 (29 May 2026)

- Added support for new connection types:
  - DealCloud
  - Knock
  - Qtanium Connect
  - Tigris

## v1.5.4 (18 May 2026)

- Added support for new connection types:
  - Autumn
  - Luma
  - Notion
  - Reo.Dev
  - ScyllaDB
- Updated connection schemas for Amazon Selling Partner, Amplitude, API, ClickHouse, CSV, Freshservice, Highspot, HTTP Enrichment, MS Ads, NetSuite, Pitchbook, and Ware2Go.

## v1.5.3 (6 May 2026)

BUG FIXES:

- Fixed misleading "partner key is required" error when using an API key with a resource that sets `organization`. The provider now verifies the organization via `/api/me` instead of the partner-only `Organization.List` endpoint.

## v1.5.2 (30 April 2026)

BUG FIXES:

- Fixed incompatibility between provider and organization discovery endpoints.

## v1.5.1 (28 April 2026)

BUG FIXES:

- Computed-only configuration fields are now stripped from connection create and update API request payloads, since they represent server-managed values.
- Preserve user-supplied sensitive connection values (passwords, service account keys, certificates, etc.) in state across Create and Update. Previously the provider stored the API's masked response, which caused "Provider produced inconsistent result after apply" errors and spurious drift on subsequent plans.

BREAKING CHANGES:

- Connection resource schemas: read-only configuration fields are now `Computed`-only and can no longer be set in user configuration. Previously these fields were generated as `Optional` + `Computed`, which silently accepted user-supplied values that the server would overwrite. Configurations that set a read-only field (e.g. `authenticated_as` on `polytomic_posthog_connection`) must remove the assignment. See the [Upgrading to v1.5.0](docs/guides/upgrading-to-v1.5.0.md) guide for the full list of affected fields.

IMPORTER:

- Connections with non-OAuth sensitive fields (e.g. PostHog API keys) are now imported with a generated sensitive Terraform input variable per missing field, instead of being skipped as OAuth connections.

## 1.4.2 (23 April 2026)

- Added support for new connection types:
  - Calendly
- Updated connection schema for BigQuery.

## 1.4.1 (17 April 2026)

- Added support for new connection types:
  - Factors AI
  - Gatsby
  - Sprig
- Updated connection schemas for Affinity and TikTok Ads.

## 1.4.0 (6 April 2026)

BREAKING CHANGES:

- `polytomic_sync` resource: The `filters`, `overrides`, and `override_fields` attributes have been redesigned. See the [Upgrading to v1.4.0](docs/guides/upgrading-to-v1.4.0.md) guide for migration instructions.
  - **`filters`**: Replaced `field_id`/`field_type` with `source { model_id, field }` reference. The server resolves field UUIDs from the source reference, so users no longer need to look up field UUIDs.
  - **`target_filters`**: New attribute for target-field filters (previously mixed into `filters` with `field_type = "Target"`). Only valid for syncs with mode `update`.
  - **`overrides`**: Replaced `field_id` with `source { model_id, field }` reference, matching the new `filters` pattern.
  - **`override_fields`**: Removed the `source` block, which was always ignored by the server. Only `target`, `override_value`, `new`, and `sync_mode` remain.
  - **`target.search_values`**: Removed. This was internal application state derived from `target.object` by the server. Users should remove any `search_values` from their `target` blocks; the server populates this automatically from the `object` field.

BUG FIXES:

- Fixed override string values causing "Invalid JSON String Value" errors due to missing JSON marshaling in response handling.

## 1.3.9 (2 April 2026)

- Added support for new connection types:
  - Amazon Selling Partner
  - Construct Connect
  - Dub
  - Greenhouse
  - Hyperline
  - Juro
  - Pinterest Ads
  - Testrail
  - Walmart Marketplace
- Configuration errors could be masked and mis-reported due to HTTP retries.

## 1.3.8 (25 Mar 2026)

- Fixed bulk sync filter values to accept any JSON type (strings, arrays, objects) via `jsonencode()`.

## 1.3.7 (18 Mar 2026)

- Added `polytomic_role` data source for looking up roles.
- Added `role_ids` field to user resource for assigning multiple/custom roles. The existing `role` field is now deprecated.

## 1.3.6 (16 Mar 2026)

- Updated documentation for `force_destroy` in connection resources.

## 1.3.5 (15 Mar 2026)

- Fixed issue with state handling when creating a new sync target.

## 1.3.4 (9 Mar 2026)

- Added support for new connection types:
  - Campfire
- Added support for accessing Databricks blob storage via SSH tunnel.
- Omit optional, unspecified fields from connection creation requests.

## 1.3.2 (25 Feb 2026)

- Added support for new connection types:
  - Clazar
  - Google Search Console
  - n8n
- Added Databricks SSH tunnel configuration fields.
- Added `use_search_api` option for HubSpot connections.

## 1.3.1 (18 Feb 2026)

- Allow non-partner keys to manage user resources
- Support importing users by email address rather than ID
- Added support for new connection types:
  - ClickHouse
  - Gorgias
  - LearnWorlds
  - Qualtrics
  - Stord

## 1.3.0 (10 Feb 2026)

- Fix type conversion error when adding new schemas to bulk syncs
- Notifications: Add global error subscribers resource

## 1.2.1 (2 Feb 2026)

- Fixed bugs with state/response drift when managing bulk syncs.

## 1.2.0 (28 Jan 2026)

- Added connection schema data source for inspecting connected data.
- Added schema primary keys resource for setting unique keys.
- Added support for new connection types:
  - Ashby
  - Chameleon
  - G2
  - Rocketlane
  - Tabs
  - Thrive
  - Ware2Go

## 1.1.1 (16 Jan 2026)

- Fixed bugs with state/response drift when managing bulk syncs.
- Fixed bugs with connection datasources erroring due to missing fields.
- Fixed normalization of deployment URL to avoid redirects.
- Added support for new connection types:
- Hightouch

## 1.1.0 (15 Jan 2026)

- Updated Polytomic Go SDK to v1.12.0.
- Added support for configuring bulk sync concurrency.
- Added support for new connection types:
  - Amplemarket
  - Appsflyer
  - Brevo
  - CallRail
  - Chili Piper
  - Cloudflare Logs
  - Docker Hub
  - Fathom
  - Fireflies AI
  - Gainsight CS
  - Gladly
  - Gmail
  - Google Slides
  - HighLevel
  - IBM Db2
  - Instantly
  - M3ter
  - MotherDuck
  - Paycor
  - Profound
  - Pylon
  - Rippling
  - Seismic
  - Showpad
  - Standard Metrics
  - TikTok Ads
  - Twilio SendGrid
  - Upfluence
  - Work OS
  - Xero
  - Zoho CRM
  - Zoho Desk

## 1.0.0 (12 Jan 2026)

- Updates from test deployments.

## 1.0.0-beta2 (03 Jan 2025)

_This is a pre-release version._

- Updated connection resources to correctly support repeated configuration properties.

## 1.0.0-beta1 (20 Dec 2024)

_This is a pre-release version._

- Rewrite to use generated Polytomic SDK.
- Model configuration is now specified as a JSON payload.
- Some connection resource and datasource names have changed.

## 0.3.41 (9 May 2024)

- Reverted breaking change from v0.3.39.

## ~~0.3.39 (29 Apr 2024)~~

**This release included a breaking change for connection configuration; v0.3.41 should be used instead.**

- Add support for overriding endpoint for Redshift Serverless connections

## 0.3.38 (7 Feb 2024)

- Fix AWS athena connections

## 0.3.37 (7 Feb 2024)

- Add support for Redshift Serverless connections

## 0.3.36 (12 Jan 2024)

- Depedency updates

## 0.3.35 (08 Dec 2023)

- Add unbounce, aws opensearch, and quickbooks connections
- Add identity endpoint data source
- Depedency updates

## 0.3.34 (10 Nov 2023)

- Depedency updates

## 0.3.33 (23 Oct 2023)

- Write sensitive values as variables when writing HCL files
- Add default values in optional connection configuration schema attributes

## 0.3.32 (19 Oct 2023)

- Add dittofeed, ironclad, mailercheck, tixr, and zoominfo connections
- Security updates

## 0.3.30 (6 Oct 2023)

- General improvements to state handling
- Make `active` attribute on sync resources required
- Add support for read-only connection attributes

## 0.3.29 (4 Oct 2023)

- State enhancements api connection resources

## 0.3.28 (4 Oct 2023)

- State enhancements for bulk syncs and connection resources

## 0.3.27 (30 Sep 2023)

- Bug fix for runAfter marshalling in bulk sync resources

## 0.3.26 (22 Sep 2023)

- Add support for Facebook Ads, Github, Slack, Google Search Console, LinkedInAds, Outreach and Datalite connections

## 0.3.25 (22 Sep 2023)

- Support float64 types in importer

## 0.3.24 (22 Sep 2023)

- Fix schedule handling in importer
- Write API key as variable in provider block

## 0.3.23 (21 Sep 2023)

- Add runAfter support for sync resources
- Importer will only optionally write API key into provider block

## 0.3.22 (15 Sep 2023)

- Fix datasource references in importer
- Add `force-destroy` attribute to datasource resources
- Skip strict connection configuration validate
- Fix bug with bulk sync schedule empty values

## 0.3.21 (29 Aug 2023)

- Documentation template tweaks

## 0.3.20 (29 Aug 2023)

- Add `force-destroy` attribute to connection resources

## 0.3.19 (28 Aug 2023)

- Various dependency updates
- Use referential identifiers in importer
- Add Facebooks Ads and DynamoDB connections

## 0.3.18 (7 July 2023)

- Add Jira connection

## 0.3.17 (16 Jun 2023)

- Improve error messaging
- Fix handling of 404s for users, organizations, and bulk syncs

## 0.3.16 (15 May 2023)

- Ensure resource names from importer are valid
- Add SSL parameter to SQL Server connection

## 0.3.15 (10 May 2023)

- Update bulk sync schema handling
- Fix handling of missing organizations

## 0.3.14 (05 Apr 2023)

- Add Asana, Databricks, Delighted, Github, Linear and LinkedIn Ads connections
- Update MySQL configuration to support change detection

## 0.3.13 (09 Mar 2023)

- Add permission resources (roles and policies)
- Add version information to importer at build time

## 0.3.12 (26 Feb 2023)

- Added Ascend and Statsig connection type

## 0.3.11 (24 Feb 2023)

- Add bingads, csv, iterable, shipbob, shopify, smartsheet, survicate, synapse, uservoice, vanilla, webhook, and zendesk connections
- Fix error updating api connections
- Ensure importer import.sh ordering is deterministic

## 0.3.10 (15 Feb 2023)

- Encode sync target configuration as JSON

## 0.3.9 (14 Feb 2023)

- Fix handling of optional organization ids
- Handle sync filters

## 0.3.8 (10 Feb 2023)

- Fix default value for polytomic url in importer

## 0.3.7 (9 Feb 2023)

- Add support for organization-level syncs, bulk syncs, and models
- Disallow deleting connections with dependent resources
- Enhanced error handling
- Add support for importing connections as data sources

## 0.3.6 (17 Jan 2023)

- Add additional example values to documentation

## 0.3.5 (16 Jan 2023)

- Re-org documentation

## 0.3.4 (16 Jan 2023)

- N/Aw

## 0.3.3 (16 Jan 2023)

- N/A

## 0.3.2 (16 Jan 2023)

- N/A

## 0.3.1 (16 Jan 2023)

- N/A

## 0.3.0 (16 Jan 2023)

- Added support for importing resources
- Updated to use new wrapped responses from API
- Added additional connection resources
- Restructured docs

## 0.2.0 (14 Dec 2022)

FEATURES

- Added support for Connection data sources
- Added support for Bulk Sync resources
- Added support for Model resources
- Added support for Model Sync resources

## 0.1.7 (26 Oct 2022)

FEATURES

- Added support for S3 Connection resources

## 0.1.6 (20 Sept 2022)

FIXES

- Corrected additional issues with case-sensitive email comparisons.
- Resources deleted outside of Terraform caused the provider to fail; they will
  now be included in the plan for creation.

## 0.1.5 (13 July 2022)

FIXES

- Email addresses are compared case insensitively when determining if a user
  needs to be replaced.

## 0.1.0 (11 July 2022)

FEATURES

- New Resources:
  - Organizations
  - Users
  - AWS Athena Connections
  - MS SQL Service Connection
