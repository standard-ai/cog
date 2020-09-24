resource "google_compute_firewall" "vault-https" {
  name    = "vault-https"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["8200"]
  }

  target_tags   = ["vault"]
  source_ranges = var.gcp_iap_https_firewall_source_ranges
}

resource "google_compute_firewall" "vault-all" {
  name    = "vault-all"
  network = "default"

  allow {
    protocol = "all"
  }

  target_tags   = ["vault"]
  source_ranges = var.gcp_iap_all_firewall_source_ranges
}
