data "template_file" "vault" {
  template = "${file("./scripts/startup.sh")}"
  vars = {
    project_id                        = var.gcp_project_id
    region                            = var.primary_region
    vault_auto_unseal_key_ring        = var.vault_auto_unseal_key_ring
    vault_auto_unseal_crypto_key_name = var.vault_auto_unseal_crypto_key_name
    storage_bucket                    = google_storage_bucket.vault.name
    vault_internal_domain             = var.vault_internal_domain
    openssl_subject                   = var.openssl_subject
  }
}

module "primary_instance" {
  source                = "./modules/compute"
  environment           = var.environment
  zone                  = var.primary_zone
  machine_type          = var.machine_type
  os_image              = var.os_image
  startup_script        = data.template_file.vault.rendered
  service_account_email = google_service_account.vault.email
  depends_on = [
    google_kms_crypto_key_iam_binding.crypto_key,
  ]
}

module "secondary_instance" {
  source                = "./modules/compute"
  environment           = var.environment
  zone                  = var.secondary_zone
  machine_type          = var.machine_type
  os_image              = var.os_image
  startup_script        = data.template_file.vault.rendered
  service_account_email = google_service_account.vault.email
  depends_on = [
    google_kms_crypto_key_iam_binding.crypto_key,
  ]
}