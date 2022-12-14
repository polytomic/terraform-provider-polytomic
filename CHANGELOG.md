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
