# Polytomic Importer

The Polytomic Terraform Importer generates Terraform HCL documents which reflect
the current in-app Polytomic configuration. This makes it suitable for
generating a backup of Polytomic configuration and determining what's changed.

## Installation

Download the appropriate binary for you architecture from the
[Releases](https://github.com/polytomic/terraform-provider-polytomic/releases)
page.

## Usage

```bash
./polytomic-importer run --api-key $POLYTOMIC_API_KEY --output terraform-imports --replace
```

## Automation

A Github Action,
[polytomic/terraform-sync](https://github.com/polytomic/terraform-sync), is
available which allows the importer to be used in a Github Action workflow to
automate updating the generated HCL files.
