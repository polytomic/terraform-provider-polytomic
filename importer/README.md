# Polytomic Importer

The Polytomic Terraform Importer generates Terraform HCL documents which reflect
the current in-app Polytomic configuration. This makes it suitable for
generating a backup of Polytomic configuration and determining what's changed.

## Installation

Download the appropriate binary for you architecture from the
[Releases](https://github.com/polytomic/terraform-provider-polytomic/releases)
page.

## Usage

### With API Key (single organization)

```bash
./polytomic-importer run --api-key $POLYTOMIC_API_KEY --output terraform-imports --replace
```

### With Partner Key (multiple organizations)

```bash
./polytomic-importer run \
  --partner-key $POLYTOMIC_PARTNER_KEY \
  --organizations "org-id-1,org-id-2" \
  --output terraform-imports \
  --replace
```

### With Deployment Key (multiple organizations)

```bash
./polytomic-importer run \
  --deployment-key $POLYTOMIC_DEPLOYMENT_KEY \
  --organizations "org-id-1,org-id-2" \
  --output terraform-imports \
  --replace
```

### Options

- `--replace`: Replace existing files (otherwise the command will fail if files exist)
- `--include-permissions`: Include role and policy resources in the import

## Authentication Options

The importer supports three authentication methods:

1. **API Key**: For single organization access. The importer will automatically
   discover your organization.

2. **Partner Key**: For multi-organization access.

   - `--partner-key`: Your Polytomic partner key
   - `--organizations`: Comma-separated list of organization IDs to import (optional)

3. **Deployment Key**: For multi-organization access (same privileges as partner key).
   - `--deployment-key`: Your Polytomic deployment key
   - `--organizations`: Comma-separated list of organization IDs to import (optional)

### Multi-Organization Mode

When using partner or deployment keys with multiple organizations:

- If `--organizations` is not specified, the importer discovers all accessible organizations
- With multiple organizations, each gets its own directory under the output path
- The directory name is based on the organization name
- Each directory contains its own set of `.tf` files and `import.sh` script

### Organization Discovery

The importer automatically discovers organizations based on the authentication method:

- **API Key**: Uses the organization associated with the API key
- **Partner/Deployment Key**: Lists all accessible organizations, optionally filtered by `--organizations`

## Automation

A Github Action,
[polytomic/terraform-sync](https://github.com/polytomic/terraform-sync), is
available which allows the importer to be used in a Github Action workflow to
automate updating the generated HCL files.
