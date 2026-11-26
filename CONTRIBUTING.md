# Contributing to terraform-provider-sanity

Thank you for your interest in contributing to the Sanity.io Terraform Provider!

## Development Requirements

- [Go](https://golang.org/doc/install) >= 1.22
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0

## Building the Provider

```bash
git clone https://github.com/valoriz/terraform-provider-sanity.git
cd terraform-provider-sanity
go build -v .
```

## Running Tests

### Unit Tests
```bash
go test -v ./...
```

## Generating Documentation

Documentation is generated automatically using `tfplugindocs`:

```bash
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name sanity
```

## Submitting Pull Requests

1. Fork the repository and create your branch from `main`.
2. Ensure all unit tests pass with `go test -v ./...`.
3. Format all code with `go fmt ./...`.
4. Submit a Pull Request describing your changes.
