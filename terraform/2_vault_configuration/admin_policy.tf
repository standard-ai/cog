resource "vault_policy" "admin" {
  name   = "admin"
  policy = <<EOT
path "${vault_mount.ssh-hostkey.path}/config/ca"
{
  capabilities = ["read", "list"]
}

path "${vault_mount.ssh.path}/config/ca"
{
  capabilities = ["read", "list"]
}
EOT
}