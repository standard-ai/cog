#!/bin/bash

export USERNAME=demouser
export PASSWORD=demopassword
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=$(cat vault_initialization.log | grep 'Initial Root Token' | cut -d : -f 2 | sed 's/ //g')

# Retrieve trusted user CA key from Vault
mkdir keys
vault read -tls-skip-verify -field=public_key ssh/config/ca > keys/trusted-user-ca-keys.pem

# Generate and sign bastion host key
ssh-keygen -q -t ed25519 -N '' -f keys/bastion
vault write -tls-skip-verify -field=signed_key ssh-hostkey/sign/ssh-hostkey \
    cert_type=host \
    public_key=@keys/bastion.pub > keys/bastion-cert.pub

# Generate and sign target host key
ssh-keygen -q -t ed25519 -N '' -f keys/target
vault write -tls-skip-verify -field=signed_key ssh-hostkey/sign/ssh-hostkey \
    cert_type=host \
    public_key=@keys/target.pub > keys/target-cert.pub

# Build the two host containers
docker-compose build --build-arg USERNAME=$USERNAME

# Run the host containers
docker-compose up -d
