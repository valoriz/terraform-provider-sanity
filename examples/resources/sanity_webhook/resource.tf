# Configure a document webhook for frontend revalidation
resource "sanity_webhook" "revalidate" {
  project_id  = sanity_project.example.id
  name        = "Frontend Static Revalidation"
  url         = "https://my-frontend.example.com/api/revalidate"
  dataset     = "production"
  api_version = "v2024-01-01"
  http_method = "POST"
  secret      = "super-secret-hmac-key"

  filter     = "_type == 'post' && !(_id in path('drafts.**'))"
  projection = "{ 'id': _id, 'slug': slug.current }"

  include_drafts   = false
  is_disabled_by_user = false
}
