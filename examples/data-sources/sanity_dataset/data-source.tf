# Look up an existing dataset
data "sanity_dataset" "prod" {
  project_id = "your-project-id"
  name       = "production"
}

output "is_dataset_public" {
  value = data.sanity_dataset.prod.acl_mode == "public"
}
