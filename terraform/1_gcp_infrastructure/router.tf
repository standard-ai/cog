module "primary_nat" {
  source = "./modules/nat"
  name   = "vault-primary"
  region = var.primary_region
}

module "secondary_nat" {
  source = "./modules/nat"
  name   = "vault-secondary"
  region = var.secondary_region
}
