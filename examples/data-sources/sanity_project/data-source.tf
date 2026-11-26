# Fetch an existing project by ID
data "sanity_project" "existing" {
  id = "your-project-id"
}

output "project_name" {
  value = data.sanity_project.existing.project_name
}
