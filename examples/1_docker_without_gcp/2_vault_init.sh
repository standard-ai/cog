#!/bin/bash

export VAULT_ADDR=http://127.0.0.1:8200

# Start HashiCorp Vault
docker-compose up -d vault

sleep 5

# Initialize HashiCorp Vault
vault operator init -key-shares=1 -key-threshold=1 > vault_initialization.log

# Unseal HashiCorp Vault
unseal_key=$(cat vault_initialization.log | grep 'Unseal Key 1' | cut -d : -f 2 | sed 's/ //g')
vault operator unseal $unseal_key
