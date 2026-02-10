# Polytomic Terraform Provider

This repository contains the [Polytomic](https://polytomic.com) Terraform
provider. This provider may be used to manage on premises deployments.

## Example Organization configuration with users

```terraform
provider "polytomic" {
  deployment_url     = "polytomic.acmeinc.com"
  deployment_api_key = "secret-key"
}

resource "polytomic_organization" "acme" {
  name = "Acme, Inc."
}

resource "polytomic_user" "acme_admin" {
  organization = polytomic_organization.acme.id
  email        = "admin@acmeinc.com"
  role         = "admin"
}
```

## Terraform Importer

See [Polytomic Importer](./importer/README.md)

## Development

If you wish to work on the provider, you'll first need
[Go](http://www.golang.org) installed on your machine (see
[Requirements](#requirements) above).

Connections can be added by editing `provider/gen/connections/connections.yaml`
and running `go generate`

To compile the provider, run `go install`. This will build the provider and put
the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

### Acceptance Tests

Acceptance tests are written to run against a real Polytomic deployment (often a local stack).

The `GNUmakefile` `testacc` target defaults `POLYTOMIC_DEPLOYMENT_URL` to `https://app.polytomic-local.com`:

```shell
make testacc
```

You can override the URL and pass through additional `go test` flags via `TESTARGS`:

```shell
POLYTOMIC_DEPLOYMENT_URL=https://app.polytomic-local.com \
  make testacc TESTARGS='-run TestAccConnectionResource -count=1 -v'
```

Environment variables required for acceptance tests:

- `POLYTOMIC_DEPLOYMENT_URL` (e.g. `https://app.polytomic-local.com`)
- One of:
  - `POLYTOMIC_API_KEY` (org-scoped)
  - `POLYTOMIC_DEPLOYMENT_KEY` (deployment/partner scoped; tests will create an org and pass `organization = ...`)

If you prefer to run an individual package directly:

```shell
TF_ACC=1 \
  POLYTOMIC_DEPLOYMENT_URL=https://app.polytomic-local.com \
  go test ./provider/... -run TestAccGlobalErrorSubscribersResource -count=1 -v
```

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.5
- [Go](https://golang.org/doc/install) >= 1.25

### Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

### Testing a pre-release version

If you wish to test a pre-release version of the provider, you can use the a
`.tfrc` file to specify the location of your local provider binary. If you
installed the provider using `go install`, create a `dev.tfrc` file with the
following contents, substituting the correct home directory path.

```terraform
provider_installation {
  dev_overrides {
    "polytomic/polytomic" = "/Users/username/go/bin"
  }

  # all the other providers, install them as usual
  direct {}
}
```

To instruct Terraform to use those settings, set the `TF_CLI_CONFIG` environment
variable.

```sh
$ TF_CLI_CONFIG_FILE=dev.tfrc terraform apply
```

## Releasing

1. Update [CHANGELOG.md](./CHANGELOG.md) with release details and date and commit.
1. Create an annotated version tag; a version tag consists of the letter `v` followed by `MAJOR.MINOR.PATCH`. For example:

   ```shell
   git tag -a v0.2.0
   ```

1. Push the tag to Github.

   ```shell
   git push origin v0.2.0
   ```

Github Actions are configured to build release tags and create a new release. Once the release has been created, the Terraform registry will pick it up within a few minutes.
