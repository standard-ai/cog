## WARNING: If you use terraform to destroy these, you won't be able to decrypt anything in the vault

resource "google_kms_key_ring" "vault" {
  name     = var.vault_auto_unseal_key_ring
  location = var.primary_region
  lifecycle {
    #    prevent_destroy = true
  }
}

resource "google_kms_crypto_key" "vault" {
  name            = var.vault_auto_unseal_crypto_key_name
  key_ring        = google_kms_key_ring.vault.id
  rotation_period = "604800s" # 60 * 60 * 24 * 7
  lifecycle {
    #    prevent_destroy = true
  }
}

resource "google_kms_crypto_key_iam_binding" "crypto_key" {
  crypto_key_id = google_kms_crypto_key.vault.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"

  members = [
    "serviceAccount:${google_service_account.vault.email}",
  ]
  lifecycle {
    #    prevent_destroy = true
  }
}
