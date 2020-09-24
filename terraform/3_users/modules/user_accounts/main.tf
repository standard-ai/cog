resource "vault_identity_entity" "user_account" {
  name     = var.username
  policies = var.policies
}

resource "vault_identity_entity_alias" "user_account" {
  name           = var.email
  mount_accessor = var.oidc_mount_accessor
  canonical_id   = vault_identity_entity.user_account.id
}

resource "vault_ssh_secret_backend_role" "user_account" {
  name                    = var.username
  backend                 = "ssh"
  key_type                = "ca"
  allow_user_certificates = true
  allowed_users           = join(",", compact(concat(list(var.username), var.unix_roles)))

  default_extensions = {
    "permit-agent-forwarding" = ""
    "permit-port-forwarding"  = ""
    "permit-pty"              = ""
    "permit-X11-forwarding"   = ""
  }

  default_user = join(",", compact(concat(list(var.username), var.unix_roles)))
  ttl          = var.ssh_sign_ttl
}
