resource "google_compute_global_address" "vault" {
  name = "vault"
}

output "load_balancer_ip_address" {
  value = google_compute_global_address.vault.address
}