#!/bin/bash

# Remove vault login
rm ~/.vault-token

# Remove SSH keys
rm -rf keys

# Stop and remove containers, networks, images, and volumes 
docker-compose down

# Remove cog configuration
rm -rf ~/.config/cog

# Remove cog stanza from ~/.ssh/config
grep 'BEGIN DEMO COG CONFIG' ~/.ssh/config
exit=$?
if [ $exit -eq 0 ] ; then
  head -n -20 ~/.ssh/config > ~/.ssh/config.tmp
  mv ~/.ssh/config.tmp ~/.ssh/config
fi
