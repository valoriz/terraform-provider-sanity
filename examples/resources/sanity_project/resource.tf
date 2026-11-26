# Create a new Sanity project
resource "sanity_project" "example" {
  project_name    = "Production Application"
  organization_id = "your-organization-id"
}

# Reference the created project ID in other resources:
# output "project_id" {
#   value = sanity_project.example.id
# }
