resource "vault_auth_backend" "userpass" {
  type = "userpass"

  tune {
    max_lease_ttl = "90000s"
  }
}