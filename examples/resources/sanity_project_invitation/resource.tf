# Invite a developer to the project
resource "sanity_project_invitation" "developer" {
  project_id = sanity_project.example.id
  email      = "qa-lead@example.com"
  role       = "editor"
}

output "invitation_id" {
  value = sanity_project_invitation.developer.id
}
