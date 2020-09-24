resource "vault_jwt_auth_backend" "oidc" {
  description        = "Google OIDC"
  path               = "oidc"
  type               = "oidc"
  oidc_discovery_url = var.oidc_discovery_url
  oidc_client_id     = var.oidc_client_id
  oidc_client_secret = var.oidc_client_secret
  default_role       = var.oidc_default_role
}

resource "vault_jwt_auth_backend_role" "oidc" {
  backend        = vault_jwt_auth_backend.oidc.path
  role_name      = "default"
  token_policies = ["default"]
  user_claim     = "email"
  oidc_scopes    = ["email"]

  bound_claims = {
    email = join(",", var.role_members_default)
  }

  bound_audiences       = var.reader_role_bound_audiences
  allowed_redirect_uris = var.reader_role_allowed_redirect_uris
  role_type             = "oidc"
}