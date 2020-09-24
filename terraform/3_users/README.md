# Terraforming of HashiCorp Vault users

This directory of terraform files creates the users in HashiCorp Vault that will be consumers of `cog`.

---

If you want to store your terraform state in the cloud, copy the `state.tf.template` file and 
uncomment the appropriate lines in `variables.tf`

```
# Optional!
cp state.tf.template state.tf
```

Copy the `terraform.tfvars.template` so that we can edit it and use it

```
cp terraform.tfvars.template terraform.tfvars
```

Edit terraform.tfvars appropriately. Uncomment the state variables if you're storing terraform state in the cloud.

Start your tunnel and log in if you need to.

```
gcloud beta compute start-iap-tunnel INSTANCE_NAME 8200 --local-host-port=127.0.0.1:8200 --zone ZONE --project PROJECT_ID
export VAULT_ADDR=https://127.0.0.1:8200
vault status -tls-skip-verify # Make sure HA is 'active', not 'standby'
vault login -tls-skip-verify
```

```
terraform init
terraform plan -out current.plan
terraform apply current.plan
```

Once this has been applied, you'll need to grab the OIDC mount:

```
vault auth list --tls-skip-verify
```

Grab the accessor for the `oidc/` path. Add it to `terraform.tfvars`:

```
oidc_mount_accessor = "ACCESSOR"
```

Copy the user_accounts.tf.template file into place and edit.

```
cp user_accounts.tf.template user_accounts.tf
```

Now re-apply terraform:

```
terraform init # You'll need to re-init to install the new module you just created
terraform plan -out current.plan
terraform apply current.plan
```

At this point, you should be able to ask vault to sign a public key for you.


```
mkdir keys
ssh-keygen -t ed25519 -N '' -f keys/personal
vault write -tls-skip-verify -field=signed_key ssh/sign/USERNAME public_key=@keys/personal.pub > keys/personal-cert.pub
ssh-keygen -L -f keys/personal-cert.pub
```

Congratulations! You've signed a certificate using Hashicorp Vault.
