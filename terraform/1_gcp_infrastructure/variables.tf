variable "gcp_project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "gcp_storage_bucket_name" {
  description = "GCP Storage Bucket Name"
  type        = string
}

variable "gcp_cog_storage_bucket_name" {
  description = "GCP Cog Storage Bucket Name"
  type        = string
}

variable "gcp_storage_bucket_location" {
  description = "GCP Storage Bucket Location"
  type        = string
  default     = "US"
}

variable "gcp_storage_bucket_storage_class" {
  description = "GCP Storage Bucket Storage Class"
  type        = string
  default     = "STANDARD"
}

variable "gcp_iap_https_firewall_source_ranges" {
  description = "GCP Firewall IP ranges (https://cloud.google.com/iap/docs/load-balancer-howto#firewall)"
  type        = tolist([string])
  default = [
    "130.211.0.0/22",
    "35.191.0.0/16",
  ]
}

variable "gcp_iap_all_firewall_source_ranges" {
  description = "GCP IAP Firewall IP ranges"
  type        = tolist([string])
  default = [
    "35.235.240.0/20",
  ]
}


variable "primary_region" {
  description = "Region for first instance of vault"
  type        = string
  default     = "us-west1"
}

variable "secondary_region" {
  description = "Region for second instance of vault"
  type        = string
  default     = "us-west2"
}

variable "service_account_id" {
  description = "Service Account ID (must be unique in the project with a minimum character length)"
  type        = string
  default     = "hashicorp-vault"
}

variable "service_account_display_name" {
  description = "Service Account Display Name"
  type        = string
  default     = "Hashicorp Vault"
}

variable "vault_auto_unseal_key_ring" {
  description = "The GCP Cloud KMS Key Ring to use for the Auto Unseal feature."
  type        = string
  default     = "hashivault"
}

variable "vault_auto_unseal_crypto_key_name" {
  description = "The GCP Cloud KMS Crypto Key to use for the Auto Unseal feature."
  type        = string
  default     = "auto_unseal"
}

variable "environment" {
  description = "Environment"
  type        = string
  default     = "dev"
}

variable "primary_zone" {
  description = "Primary Zone"
  type        = string
  default     = "us-west1-a"
}

variable "secondary_zone" {
  description = "Secondary Zone"
  type        = string
  default     = "us-west2-a"
}

variable "machine_type" {
  description = "Machine Type"
  type        = string
  default     = "n1-standard-2"
}

variable "os_image" {
  description = "OS Image"
  type        = string
  default     = "ubuntu-1804-lts"
}

variable "vault_internal_domain" {
  description = "Internal only domain name"
  type        = string
}

variable "vault_external_domain" {
  description = "External domain name"
  type        = string
}

variable "oauth2_client_id" {
  description = "OAuth2 Client ID"
  type        = string
}

variable "oauth2_client_secret" {
  description = "OAuth2 Client Secret"
  type        = string
}

variable "vault_iam_members" {
  description = "Vault IAM members"
  type        = tolist([string])
}

variable "openssl_subject" {
  description = "Subject string for self-signed certificate generation"
  type        = string
}

variable "iam_roles_owners" {
  description = "Members who get roles/owner"
  type        = tolist([string])
}

variable "iap_email_address" {
  description = "Email address for IAP Branding"
  type        = string
}