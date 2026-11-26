# Create a production dataset
resource "sanity_dataset" "production" {
  project_id = sanity_project.example.id
  name       = "production"
  acl_mode   = "public"
}

# Create a staging dataset
resource "sanity_dataset" "staging" {
  project_id = sanity_project.example.id
  name       = "staging"
  acl_mode   = "public"
}
