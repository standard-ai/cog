#!/bin/bash

export VAULT_ADDR=https://127.0.0.1:8200
export USERNAME=$(grep user_account_ ../../terraform/3_users/user_accounts.tf | awk '{print $2}' | sed 's/"user_account_//g' | sed 's/"//g')
export COG_VAULT_ADDR=https://$(grep vault_external_domain ../../terraform/1_gcp_infrastructure/terraform.tfvars | awk '{print $3}' | sed 's/"//g')
export COG_VAULT_PROXY_HOST=$(grep vault_internal_domain ../../terraform/1_gcp_infrastructure/terraform.tfvars | awk '{print $3}' | sed 's/"//g')
export COG_IAP_CLIENT_ID=$(grep oauth2_client_id ../../terraform/1_gcp_infrastructure/terraform.tfvars | awk '{print $3}' | sed 's/"//g')
export SERVICE_ACCOUNT_ID=$(grep service_account_id ../../terraform/1_gcp_infrastructure/terraform.tfvars | awk '{print $3}' | sed 's/"//g')
export PROJECT=$(grep gcp_project_id ../../terraform/1_gcp_infrastructure/terraform.tfvars | awk '{print $3}' | sed 's/"//g')
export COG_IAP_SERVICE_ACCT="${SERVICE_ACCOUNT_ID}@${PROJECT}.iam.gserviceaccount.com"
export COG_GCS_BUCKET=$(grep gcp_cog_storage_bucket_name ../../terraform/1_gcp_infrastructure/terraform.tfvars | awk '{print $3}' | sed 's/"//g')

# Build cog binary
cd ../..
sed -i '' -e "s#^\(default_vaultAddress\)=.*#\1=${COG_VAULT_ADDR}#" scripts/build.sh
sed -i '' -e "s#^\(default_vaultProxyHost\)=.*#\1=${COG_VAULT_PROXY_HOST}#" scripts/build.sh
sed -i '' -e "s#^\(default_vaultIAPServiceAccount\)=.*#\1=${COG_IAP_SERVICE_ACCT}#" scripts/build.sh
sed -i '' -e "s#^\(default_vaultIAPClientID\)=.*#\1=${COG_IAP_CLIENT_ID}#" scripts/build.sh
sed -i '' -e "s#^\(default_gcsBucket\)=.*#\1=${COG_GCS_BUCKET}#" scripts/build.sh
sed -i '' -e "s#^\(default_gcsFilename\)=.*#\1=inventory.yaml#" scripts/build.sh
sed -i '' -e "s#^\(default_binaryGCSBucket\)=.*#\1=${COG_GCS_BUCKET}#" scripts/build.sh
sed -i '' -e "s#^\(default_binaryGCSPath\)=.*#\1=bin/#" scripts/build.sh
make
make install

# Create ~/.config/cog/inventory.yaml
cat > inventory.yaml <<EOF
map:
- bastions:
  - cog_bastion
  globs:
  - '*'
  hosts:
  - cog_target
  - cog_bastion
  ssh_ca: ssh
  ssh_group: global
EOF

# Copy inventory.yaml to GCS
gsutil cp inventory.yaml gs://${COG_GCS_BUCKET}/inventory.yaml
rm inventory.yaml

cat > ssh_config <<EOF
## BEGIN COG CONFIGURATION
Host cog_bastion
    HostName 127.0.0.1
    Port 2222
    ServerAliveInterval 60
    ServerAliveCountMax 5
    ControlPath ~/.ssh/a-%C
    ControlMaster auto
    ControlPersist 30m
    ForwardAgent yes
    UserKnownHostsFile ~/.config/cog/known_hosts

Host cog_target
    User ubuntu
    ProxyCommand cog sshproxy -W %h:%p
    UserKnownHostsFile ~/.config/cog/known_hosts
## END COG CONFIGURATION
EOF

# Copy ssh_config to GCS
gsutil cp ssh_config gs://${COG_GCS_BUCKET}/ssh_config
rm ssh_config

echo "" > known_hosts

# Copy empty known_hosts to GCS
gsutil cp known_hosts gs://${COG_GCS_BUCKET}/known_hosts

# Initialize cog
cog -v init

# Use `cog vault-proxy run` to create good known_hosts
echo "@cert-authority cog_target $(cog vault-proxy run vault -- read -field=public_key ssh-hostkey/config/ca)" > known_hosts
echo "@cert-authority [127.0.0.1]:2222 $(cog vault-proxy run vault -- read -field=public_key ssh-hostkey/config/ca)" >> known_hosts

# Copy known_hosts to GCS
gsutil cp known_hosts gs://${COG_GCS_BUCKET}/known_hosts
rm known_hosts

# Re-initialize cog to pull good known_hosts down
cog -v init
