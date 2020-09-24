resource "google_compute_health_check" "vault" {
  name               = "vault"
  timeout_sec        = 5
  check_interval_sec = 5

  https_health_check {
    host         = var.vault_internal_domain
    port         = 8200
    request_path = "/v1/sys/health"
  }
}

resource "google_compute_backend_service" "vault" {
  name        = "vault"
  port_name   = "vault"
  protocol    = "HTTPS"
  timeout_sec = 300

  backend {
    group = module.primary_instance.instance_group_self_link
  }

  backend {
    group = module.secondary_instance.instance_group_self_link
  }

  health_checks = [
    google_compute_health_check.vault.self_link,
  ]

  iap {
    oauth2_client_id     = var.oauth2_client_id
    oauth2_client_secret = var.oauth2_client_secret
  }

}


resource "google_compute_url_map" "vault" {
  name            = "vault"
  default_service = google_compute_backend_service.vault.self_link
}

resource "google_compute_managed_ssl_certificate" "vault" {
  provider = google-beta

  name = "vault"

  managed {
    domains = [var.vault_external_domain]
  }
}

resource "google_compute_target_https_proxy" "vault" {
  name             = "vault-https-proxy"
  url_map          = google_compute_url_map.vault.self_link
  ssl_certificates = [google_compute_managed_ssl_certificate.vault.self_link]
}

resource "google_compute_global_forwarding_rule" "vault" {
  name       = "vault"
  ip_address = google_compute_global_address.vault.address
  target     = google_compute_target_https_proxy.vault.self_link
  port_range = "443"
}

resource "google_project_iam_binding" "project" {
  project = var.gcp_project_id
  role    = "roles/iap.httpsResourceAccessor"
  members = concat(var.vault_iam_members, ["serviceAccount:${google_service_account.vault.email}"])
}

resource "google_project_iam_binding" "iap" {
  project = var.gcp_project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  members = concat(var.vault_iam_members, ["serviceAccount:${google_service_account.vault.email}"])
}
