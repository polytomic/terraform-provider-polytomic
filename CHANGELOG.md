## 0.3.7 (10 Feb 2023)
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
