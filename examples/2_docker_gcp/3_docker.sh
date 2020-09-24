#!/bin/bash

export USERNAME=$(grep user_account_ ../../terraform/3_users/user_accounts.tf | awk '{print $2}' | sed 's/"user_account_//g' | sed 's/"//g')

# Use `cog vault-proxy run` to retrieve trusted user CA key from Vault
mkdir keys

../../build/cog vault-proxy run vault -- read -field=public_key ssh/config/ca > keys/trusted-user-ca-keys.pem

# Use `cog vault-proxy run` to generate and sign bastion host key
ssh-keygen -q -t ed25519 -N '' -f keys/bastion
../../build/cog vault-proxy run vault -- write -field=signed_key ssh-hostkey/sign/ssh-hostkey \
    cert_type=host \
    public_key=@keys/bastion.pub > keys/bastion-cert.pub

# Use `cog vault-proxy run` to generate and sign target host key
ssh-keygen -q -t ed25519 -N '' -f keys/target
../../build/cog vault-proxy run vault -- write -field=signed_key ssh-hostkey/sign/ssh-hostkey \
    cert_type=host \
    public_key=@keys/target.pub > keys/target-cert.pub

# Build the two host containers
docker-compose build --build-arg USERNAME=$USERNAME

# Run the host containers
docker-compose up -d
