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

## Development

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.18

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
