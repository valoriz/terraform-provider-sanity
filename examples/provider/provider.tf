terraform {
  required_providers {
    sanity = {
      source  = "valoriz/sanity"
      version = "~> 1.0"
    }
  }
}

# The provider reads the SANITY_TOKEN environment variable by default.
# You can also pass it explicitly if needed:
# provider "sanity" {
#   api_token = var.sanity_token
# }
provider "sanity" {}
