variable "username" {
  description = "Unix username"
}

variable "email" {
  description = "Official email address"
}

variable "unix_roles" {
  description = "Array of roles the user can access"
  default     = []
}

variable "policies" {
  description = "Vault policies"
  default     = []
}

variable "oidc_mount_accessor" {
  default = ""
}

variable "ssh_sign_ttl" {
  default = "86400"
}
