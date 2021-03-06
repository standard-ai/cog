#!/bin/bash

# Remove any existing cog SSH keys
rm -rf keys terraform/keys

# Remove any existing terraform directory
rm -rf terraform/.terraform

# Stop and remove containers, networks, images, and volumes 
docker-compose down

# Remove any persistent vault data
rm -rf vault/{data,file,logs} VAULT_PASSWORD vault_initialization.log

# Remove cog configuration
rm -rf ~/.config/cog

# Remote vault token credentials
rm -rf ~/.vault_token

# Remove cog stanza from ~/.ssh/config
grep 'BEGIN DEMO COG CONFIG' ~/.ssh/config
exit=$?
if [ $exit -eq 0 ] ; then
  head -n -20 ~/.ssh/config > ~/.ssh/config.tmp
  mv ~/.ssh/config.tmp ~/.ssh/config
fi
