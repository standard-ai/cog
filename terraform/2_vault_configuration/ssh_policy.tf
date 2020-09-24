resource "vault_policy" "ssh" {
  name = "ssh"

  policy = <<EOT
path "${vault_mount.ssh.path}/sign/{{identity.entity.name}}" {
  capabilities = ["update"]
}
EOT
}