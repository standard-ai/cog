# Terraforming of GCP Infrastructure

This directory of terraform files creates the GCP infrastructure needed to run Hashicorp Vault for cog.

The terraform resources created:

```
google_compute_backend_service
google_compute_firewall
google_compute_global_address
google_compute_global_forwarding_rule
google_compute_health_check
google_compute_managed_ssl_certificate
google_compute_target_https_proxy
google_compute_url_map
google_iap_brand
google_kms_crypto_key
google_kms_crypto_key_iam_binding
google_kms_key_ring
google_project_iam_binding
google_project_iam_member
google_project_service
google_service_account
google_storage_bucket
```

---

Prior to terraforming, we must set up OAuth2, which is not terraformable.

In your browser, go to https://console.cloud.google.com/apis/credentials

`+ CREATE CREDENTIALS` -> OAuth client ID -> Application type: Web application

Choose an appropriate Name, like "My Vault Demo"

We need to add URIs to the `Authorized redirect URIs` section.

If your external domain is `vault.example.com`, the URIs entered should look like this:

```
https://vault.example.gom
https://vault.example.gom/oidc/callback
https://vault.example.gom/ui/vault/auth/oidc/oidc/callback
http://localhost:8250/oidc/callback
```

Click `Create`. Copy your Client ID and Your Client Secret.

Go back and edit the OAuth2 client you just created.

Add one more URI. This one must contain the client ID that you copied above:

```
https://iap.googleapis.com/v1/oauth/clientIds/YOURCLIENTID.apps.googleusercontent.com:handleRedirect
```

For example:

```
https://iap.googleapis.com/v1/oauth/clientIds/012345678901-ab12cd34fg45ivhwjekd2jrngtktjd99.apps.googleusercontent.com:handleRedirect
```

This will will allow the Identity Aware Proxy to work properly in the future.

On to the terraforming!

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

Edit `terraform.tfvars` appropriately. Uncomment the state variables if you're storing terraform state in the cloud. This project assumes that the GCP project already exists. If it does not, create it with `gcloud`. See `gcloud projects create --help` for details.

Initialize Terraform

```
terraform init
```

We need the public IP address so that we can create a public facing DNS entry for it. This is a pre-requisite
for Google to provision an SSL cert. 

```
terraform plan -out current.plan -target=google_compute_global_address.vault
terraform apply current.plan
```

Create a public facing DNS entry for the IP marked as `load_balancer_ip_address`. The DNS entry must match
your `vault_external_domain` in `terraform.tfvars`. When provisioning vault, you must also have a valid DNS 
entry for `vault_internal_domain`. Point this at `127.0.0.1`.

One way to do that is with the `gcloud` command line:

```
gcloud dns record-sets transaction start --project=DNS_PROJECT --zone=DNS_ZONE
gcloud dns record-sets transaction add LOAD_BALANCER_IP_ADDRESS --name=VAULT_EXTERNAL_DOMAIN --type=A --ttl=300
gcloud dns record-sets transaction add 127.0.0.1 --name=VAULT_INTERNAL_DOMAIN --type=A --ttl=300
gcloud dns record-sets transaction execute --project=DNS_PROJECT --zone=DNS_ZONE
```

Now you can create the rest of the infrastructure:

```
terraform plan -out current.plan
terraform apply current.plan
```

To verify a bit that your instances are up:

```
gcloud compute instances list --project=YOUR_PROJECT_ID
```