resource "vault_policy" "ssh" {
  name = "ssh-${var.username}"

  policy = <<EOT
path "${var.policy_path}" {
  capabilities = ["update"]
}
EOT
}

resource "vault_generic_endpoint" "user_account" {
  path                 = "auth/userpass/users/${var.username}"
  ignore_absent_fields = true

  data_json = <<EOT
{
  "policies": ${jsonencode(var.policies)},
  "password": "${var.password}"
}
EOT
}

resource "vault_ssh_secret_backend_role" "user_account" {
  name                    = var.username
  backend                 = "ssh"
  key_type                = "ca"
  allow_user_certificates = true
  allowed_users           = join(",", compact(concat(tolist([var.username]), var.unix_roles)))

  default_extensions = {
    "permit-agent-forwarding" = ""
    "permit-port-forwarding"  = ""
    "permit-pty"              = ""
    "permit-X11-forwarding"   = ""
  }

  default_user = join(",", compact(concat(tolist([var.username]), var.unix_roles)))
  ttl          = var.ssh_sign_ttl
}
