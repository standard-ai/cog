resource "vault_mount" "ssh-hostkey" {
  type                  = "ssh"
  path                  = "ssh-hostkey"
  description           = "SSH Host Signer"
  max_lease_ttl_seconds = "315360000"
}

resource "vault_ssh_secret_backend_ca" "ssh-hostkey" {
  backend     = vault_mount.ssh-hostkey.path
  private_key = file("keys/hostkey")
  public_key  = file("keys/hostkey.pub")
}

resource "vault_ssh_secret_backend_role" "ssh-hostkey" {
  name                    = "ssh-hostkey"
  backend                 = vault_mount.ssh-hostkey.path
  key_type                = "ca"
  ttl                     = "315360000"
  allow_host_certificates = "true"
  allowed_domains         = "nonstandard.ai"
  allow_subdomains        = true
}

resource "vault_policy" "ssh-hostkey" {
  name   = "ssh-hostkey"
  policy = <<EOT
path "${vault_mount.ssh-hostkey.path}/sign/${vault_ssh_secret_backend_role.ssh-hostkey.name}"
{
  capabilities = ["create", "read", "update", "list"]
}
EOT
}