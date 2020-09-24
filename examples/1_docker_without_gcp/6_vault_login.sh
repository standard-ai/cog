#!/bin/bash

export VAULT_ADDR=http://127.0.0.1:8200
export USERNAME=demouser
export PASSWORD=$(cat VAULT_PASSWORD)

# Log in to HashiCorp Vault
vault login -method=userpass username=${USERNAME} password=${PASSWORD}