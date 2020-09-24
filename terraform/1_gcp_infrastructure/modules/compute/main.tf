resource "google_compute_instance" "vault" {
  allow_stopping_for_update = true
  name                      = "vault0001-${var.environment}-${var.zone}"
  hostname                  = "vault0001.${var.environment}.${var.zone}"
  machine_type              = var.machine_type
  zone                      = var.zone

  metadata_startup_script = var.startup_script

  tags = [
    "vault",
  ]

  boot_disk {
    initialize_params {
      image = var.os_image
    }
  }

  network_interface {
    network = "default"
  }

  service_account {
    email  = var.service_account_email
    scopes = ["storage-rw", "cloud-platform"]
  }

}

resource "google_compute_instance_group" "vault" {
  name      = "vault"
  zone      = var.zone
  instances = google_compute_instance.vault.*.self_link

  named_port {
    name = "vault"
    port = "8200"
  }
}

output "instance_group_self_link" {
  value = google_compute_instance_group.vault.self_link
}