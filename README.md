# Terraform Provider for Sanity.io

[![CI](https://github.com/valoriz/terraform-provider-sanity/actions/workflows/ci.yml/badge.svg)](https://github.com/valoriz/terraform-provider-sanity/actions/workflows/ci.yml)
[![Terraform Registry](https://img.shields.io/badge/Terraform%20Registry-v1.0.0-blue)](https://registry.terraform.io/providers/valoriz/sanity/latest)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](LICENSE)

A Terraform provider for managing [Sanity.io](https://www.sanity.io) resources. This provider allows teams to automate Sanity infrastructure—such as projects, datasets, CORS origins, document webhooks, and team member invitations—using Infrastructure as Code.

---

## Supported Resources and Data Sources

| Type | Name | Description |
|---|---|---|
| **Resource** | [`sanity_project`](docs/resources/project.md) | Creates and manages Sanity projects. |
| **Resource** | [`sanity_dataset`](docs/resources/dataset.md) | Creates and manages datasets with public or private access control. |
| **Resource** | [`sanity_cors_origin`](docs/resources/cors_origin.md) | Configures allowed CORS origins and credential rules. |
| **Resource** | [`sanity_webhook`](docs/resources/webhook.md) | Configures document webhooks, filters, projections, and signing secrets. |
| **Resource** | [`sanity_project_invitation`](docs/resources/project_invitation.md) | Sends email invitations to add team members with assigned roles. |
| **Data Source** | [`sanity_project`](docs/data-sources/project.md) | Retrieves information about an existing Sanity project. |
| **Data Source** | [`sanity_dataset`](docs/data-sources/dataset.md) | Retrieves information about an existing dataset. |

---

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://go.dev/doc/install) >= 1.22 (if building from source)

---

## Getting Started

### 1. Configure the Provider

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    sanity = {
      source  = "valoriz/sanity"
      version = "~> 1.0"
    }
  }
}

provider "sanity" {}
```

### 2. Authentication

The provider needs a Sanity API token to make requests. We recommend passing the token using the `SANITY_TOKEN` environment variable:

```bash
export SANITY_TOKEN="your-sanity-token"
```

You can also pass it directly in the provider block:

```hcl
provider "sanity" {
  api_token = var.sanity_token
}
```

#### Token Permissions
- **To manage datasets, CORS, webhooks, and members:** Use an API token with `Developer` and `Access Manager` permissions (created in Sanity under **Project Settings > API > Tokens**).
- **To create or rename projects:** Use a token with `Administrator` permissions (created under **Organization Settings > API > Tokens** with `Blueprints` and `Manage Access` enabled).

---

## Complete Example

Here is an example that creates a Sanity project with environments (`production` and `qa`), CORS rules, a deploy webhook, and an invitation:

```hcl
# Create the Sanity project
resource "sanity_project" "main" {
  project_name    = "My Application"
  organization_id = "your-organization-id"
}

# Create production and QA datasets
resource "sanity_dataset" "production" {
  project_id = sanity_project.main.id
  name       = "production"
  acl_mode   = "public"
}

resource "sanity_dataset" "qa" {
  project_id = sanity_project.main.id
  name       = "qa"
  acl_mode   = "public"
}

# Allow web origins to connect to the Sanity API
resource "sanity_cors_origin" "production_web" {
  project_id        = sanity_project.main.id
  origin            = "https://example.com"
  allow_credentials = false
}

resource "sanity_cors_origin" "local_studio" {
  project_id        = sanity_project.main.id
  origin            = "http://localhost:3333"
  allow_credentials = true
}

# Configure a deploy webhook for production builds
resource "sanity_webhook" "production_deploy" {
  project_id          = sanity_project.main.id
  name                = "Production Deploy Hook"
  url                 = "https://example.com/api/webhooks/revalidate"
  dataset             = sanity_dataset.production.name
  api_version         = "v2024-01-01"
  http_method         = "POST"
  secret              = "webhook-signing-secret"
  include_drafts      = false
  is_disabled_by_user = false
}

# Invite a developer to the project
resource "sanity_project_invitation" "developer" {
  project_id = sanity_project.main.id
  email      = "developer@example.com"
  role       = "editor"
}
```

---

## Local Development

### Build from source
```bash
git clone https://github.com/valoriz/terraform-provider-sanity.git
cd terraform-provider-sanity
go build -v .
```

### Run tests
```bash
go test -v ./...
```

### Check code format
```bash
gofmt -s -d .
```

---

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on code standards and submitting pull requests.

## Security

To report a security issue, please review our [Security Policy](SECURITY.md).

## License

This project is licensed under the [Mozilla Public License Version 2.0](LICENSE).
