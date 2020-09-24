# Terraforming of HashiCorp Vault

This directory of terraform files creates the configuration HashiCorp Vault needs to run `cog`.

The terraform resources created:

```
 vault_mount  ssh
 vault_mount  ssh-hostkey
 vault_policy  admin
 vault_policy  ssh
 vault_policy  ssh-hostkey
 vault_ssh_secret_backend_ca  ssh
 vault_ssh_secret_backend_ca  ssh-hostkey
 vault_ssh_secret_backend_role  ssh-hostkey
```

---
First, we need to initialize Vault. List the instances:

```
gcloud compute instances list --project=PROJECT_ID
```

Now start a tunnel to one of the servers (it doesn't matter which, at the moment):

```
gcloud beta compute start-iap-tunnel INSTANCE_NAME 8200 --local-host-port=127.0.0.1:8200 --zone ZONE --project PROJECT_ID
```

If it doesn't say:

```
Testing if tunnel connection works.
Listening on port [8200].
```

Something has gone wrong.

Let's initialize vault. Set an environment variable pointing at localhost:

```
export VAULT_ADDR=https://localhost:8200
```

And show that Vault is running:

```
vault status -tls-skip-verify
```

Now initialize it:

```
vault operator init -tls-skip-verify > vault_initialization.log
```

Store the Recovery keys and Initial Root Token somewhere safe. 

Get the status of the vault server:

```
vault status -tls-skip-verify
```

If the HA Mode says `standby` instead of active, you'll need to change your `gcloud beta compute start-iap-tunnel`
to point at the other compute instance. Do so now.

Log in to Vault with the root token that was created in the initialization step.

```
vault login -tls-skip-verify
```

We'll need SSH hostkeys and signing keys to terraform properly.

```
mkdir keys
ssh-keygen -t ed25519 -N '' -f keys/hostkey
ssh-keygen -t ed25519 -N '' -f keys/signkey
```

Now you can continue with the rest of the terraforming.

---

If you want to store your terraform state in the cloud, copy the `state.tf.template` file and 
uncomment the appropriate lines in `variables.tf`

```
# Optional!
cp state.tf.template state.tf
```

```
terraform init
terraform plan -out current.plan
terraform apply current.plan
```
