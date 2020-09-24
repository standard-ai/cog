variable "username" {
  description = "Unix username"
}

variable "password" {
  description = "Cleartext password"
}

variable "unix_roles" {
  description = "Array of roles the user can access"
  default     = []
}

variable "policies" {
  description = "Vault policies"
  default     = []
}

variable "ssh_sign_ttl" {
  default = "86400"
}

variable "policy_path" {}