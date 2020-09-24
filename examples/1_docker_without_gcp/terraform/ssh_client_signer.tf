resource "vault_mount" "ssh" {
  type                  = "ssh"
  path                  = "ssh"
  description           = "SSH Client Signer"
  max_lease_ttl_seconds = "315360000"
}

resource "vault_ssh_secret_backend_ca" "ssh" {
  backend     = vault_mount.ssh.path
  private_key = file("keys/signkey")
  public_key  = file("keys/signkey.pub")
}