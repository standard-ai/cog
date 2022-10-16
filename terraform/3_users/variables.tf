variable "oidc_discovery_url" {
  type    = string
  default = "https://accounts.google.com"
}

variable "oidc_client_id" {
  type = string
}

variable "oidc_client_secret" {
  type = string
}

variable "oidc_default_role" {
  type    = string
  default = "default"
}

variable "reader_role_bound_audiences" {
  type = tolist([string])
}

variable "reader_role_allowed_redirect_uris" {
  type = tolist([string])
}

variable "role_members_default" {
  type = tolist([string])
}

variable "oidc_mount_accessor" {
  default = ""
}