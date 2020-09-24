resource "google_storage_bucket" "vault" {
  name          = var.gcp_storage_bucket_name
  location      = var.gcp_storage_bucket_location
  storage_class = var.gcp_storage_bucket_storage_class
}

resource "google_storage_bucket" "cog" {
  name          = var.gcp_cog_storage_bucket_name
  location      = var.gcp_storage_bucket_location
  storage_class = var.gcp_storage_bucket_storage_class
}