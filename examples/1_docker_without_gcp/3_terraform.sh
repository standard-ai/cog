#!/bin/bash

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=$(cat vault_initialization.log | grep 'Initial Root Token' | cut -d : -f 2 | sed 's/ //g')
export USERNAME=demouser
export PASSWORD=$(cat /dev/urandom | env LC_CTYPE=C tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
echo "$PASSWORD" > ./VAULT_PASSWORD

cd terraform/

# Create SSH host key and signing key
mkdir keys
ssh-keygen -q -t ed25519 -N '' -f keys/hostkey
ssh-keygen -q -t ed25519 -N '' -f keys/signkey

# Create a sample user account terraform file
sed "s/USERNAME/$USERNAME/g" user_accounts.tf.template | sed "s/PASSWORD/$PASSWORD/g" > user_accounts.tf

# Run terraform
terraform init
terraform plan -out current.plan
terraform apply current.plan
