# Add frontend application origin
resource "sanity_cors_origin" "frontend" {
  project_id        = sanity_project.example.id
  origin            = "https://my-app.example.com"
  allow_credentials = false
}

# Add local development studio origin with credentials allowed
resource "sanity_cors_origin" "local_studio" {
  project_id        = sanity_project.example.id
  origin            = "http://localhost:3333"
  allow_credentials = true
}
