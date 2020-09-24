resource "google_compute_router" "vault" {
  name    = var.name
  region  = var.region
  network = "default"
}

resource "google_compute_router_nat" "vault" {
  name                               = var.name
  router                             = google_compute_router.vault.name
  region                             = google_compute_router.vault.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}