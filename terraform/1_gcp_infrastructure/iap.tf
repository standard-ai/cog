resource "google_project_service" "vault-iap" {
  project = data.google_project.vault.project_id
  service = "iap.googleapis.com"
}

#resource "google_iap_brand" "vault" {
#  support_email     = var.iap_email_address
#  application_title = "Cloud IAP protected Hashicorp Vault"
#  project           = google_project_service.vault-iap.project
#}