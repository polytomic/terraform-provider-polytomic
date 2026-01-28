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
