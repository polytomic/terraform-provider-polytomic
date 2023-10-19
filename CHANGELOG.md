## 0.3.32 (19 Oct 2023)
* Add dittofeed, ironclad, mailercheck, tixr, and zoominfo connections
* Security updates

## 0.3.30 (6 Oct 2023)
* General improvements to state handling
* Make `active` attribute on sync resources required
* Add support for read-only connection attributes

## 0.3.29 (4 Oct 2023)
* State enhancements api connection resources

## 0.3.28 (4 Oct 2023)
* State enhancements for bulk syncs and connection resources

## 0.3.27 (30 Sep 2023)
* Bug fix for runAfter marshalling in bulk sync resources

## 0.3.26 (22 Sep 2023)
* Add support for Facebook Ads, Github, Slack, Google Search Console, LinkedInAds, Outreach and Datalite connections


## 0.3.25 (22 Sep 2023)
* Support float64 types in importer

## 0.3.24 (22 Sep 2023)
* Fix schedule handling in importer
* Write API key as variable in provider block

## 0.3.23 (21 Sep 2023)
* Add runAfter support for sync resources
* Importer will only optionally write API key into provider block

## 0.3.22 (15 Sep 2023)
* Fix datasource references in importer
* Add `force-destroy` attribute to datasource resources
* Skip strict connection configuration validate
* Fix bug with bulk sync schedule empty values

## 0.3.21 (29 Aug 2023)
* Documentation template tweaks

## 0.3.20 (29 Aug 2023)
* Add `force-destroy` attribute to connection resources

## 0.3.19 (28 Aug 2023)
* Various dependency updates
* Use referential identifiers in importer
* Add Facebooks Ads and DynamoDB connections

## 0.3.18 (7 July 2023)
* Add Jira connection

## 0.3.17 (16 Jun 2023)
* Improve error messaging
* Fix handling of 404s for users, organizations, and bulk syncs

## 0.3.16 (15 May 2023)
* Ensure resource names from importer are valid
* Add SSL parameter to SQL Server connection


## 0.3.15 (10 May 2023)
* Update bulk sync schema handling
* Fix handling of missing organizations

## 0.3.14 (05 Apr 2023)
* Add Asana, Databricks, Delighted, Github, Linear and LinkedIn Ads connections
* Update MySQL configuration to support change detection


## 0.3.13 (09 Mar 2023)
* Add permission resources (roles and policies)
* Add version information to importer at build time

## 0.3.12 (26 Feb 2023)
* Added Ascend and Statsig connection type

## 0.3.11 (24 Feb 2023)
 * Add bingads, csv, iterable, shipbob, shopify, smartsheet, survicate, synapse, uservoice, vanilla, webhook, and zendesk connections
 * Fix error updating api connections
 * Ensure importer import.sh ordering is deterministic

## 0.3.10 (15 Feb 2023)
* Encode sync target configuration as JSON

## 0.3.9 (14 Feb 2023)
* Fix handling of optional organization ids
* Handle sync filters


## 0.3.8 (10 Feb 2023)
* Fix default value for polytomic url in importer

## 0.3.7 (9 Feb 2023)
* Add support for organization-level syncs, bulk syncs, and models
* Disallow deleting connections with dependent resources
* Enhanced error handling
* Add support for importing connections as data sources

## 0.3.6 (17 Jan 2023)
* Add additional example values to documentation
## 0.3.5 (16 Jan 2023)
* Re-org documentation
## 0.3.4 (16 Jan 2023)
* N/Aw
## 0.3.3 (16 Jan 2023)
* N/A
## 0.3.2 (16 Jan 2023)
* N/A

## 0.3.1 (16 Jan 2023)
* N/A

## 0.3.0 (16 Jan 2023)
* Added support for importing resources
* Updated to use new wrapped responses from API
* Added additional connection resources
* Restructured docs

## 0.2.0 (14 Dec 2022)

FEATURES

* Added support for Connection data sources
* Added support for Bulk Sync resources
* Added support for Model resources
* Added support for Model Sync resources

## 0.1.7 (26 Oct 2022)

FEATURES

* Added support for S3 Connection resources

## 0.1.6 (20 Sept 2022)

FIXES

* Corrected additional issues with case-sensitive email comparisons.
* Resources deleted outside of Terraform caused the provider to fail; they will
  now be included in the plan for creation.

## 0.1.5 (13 July 2022)

FIXES

* Email addresses are compared case insensitively when determining if a user
  needs to be replaced.

## 0.1.0 (11 July 2022)

FEATURES

* New Resources:
    - Organizations
    - Users
    - AWS Athena Connections
    - MS SQL Service Connection
