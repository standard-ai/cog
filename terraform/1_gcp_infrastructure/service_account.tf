resource "google_service_account" "vault" {
  account_id   = var.service_account_id
  display_name = var.service_account_display_name
}

resource "google_project_iam_member" "vault" {
  role   = "roles/editor"
  member = "serviceAccount:${google_service_account.vault.email}"
}

resource "google_project_iam_member" "vault-storage" {
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.vault.email}"
}
