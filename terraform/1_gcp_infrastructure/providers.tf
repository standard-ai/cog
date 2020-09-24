provider "google" {
  version = "~> 3.35.0"
  project = var.gcp_project_id
}

provider "google-beta" {
  version = "~> 3.35.0"
  project = var.gcp_project_id
}