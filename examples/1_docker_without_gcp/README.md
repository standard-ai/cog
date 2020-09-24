# Demoing cog with local infrastructure

## Overview

This demo uses terraform, vault, and docker containers (via docker-compose) to create a fully running
example of `cog`.

## Prerequisites

* [HashiCorp Terraform 0.13](https://www.terraform.io/downloads.html)
* [HashiCorp Vault](https://www.vaultproject.io/downloads)
* [Docker](https://www.docker.com/products/docker-desktop)
* [Golang](https://golang.org/dl/)

## TL;DR

```
./build.sh
ssh ubuntu@cog_target
```

## The scripts

These scripts should be run in order. The [build.sh](build.sh) script runs scripts `1_clear.sh` through `6_vault_login.sh`.

### 1_clear.sh

* Remove any existing cog SSH keys
* Remove any existing terraform directory
* Stop and remove containers, networks, images, and volumes
* Remove any persistent vault data
* Remove cog configuration
* Remote vault token credentials
* Remove cog stanza from ~/.ssh/config

### 2_vault_init.sh

* Start HashiCorp Vault
* Initialize HashiCorp Vault
* Unseal HashiCorp Vault

### 3_terraform.sh

* Create SSH host key and signing key
* Create a sample user account terraform file
* Run terraform

### 4_docker.sh

* Retrieve trusted user CA key from Vault
* Generate and sign bastion host key
* Generate and sign target host key
* Build the two host containers
* Run the host containers

### 5_build_cog.sh

* Create ~/.config/cog/cog.yaml
* Create ~/.config/cog/inventory.yaml
* Add CA hostkey to ~/.config/cog/known_hosts
* Add configuration to ~/.ssh/config
* Build cog binary

### 6_vault_login.sh

* Log in to HashiCorp Vault

### 7_run_cog.sh

This script just runs cog to SSH into the target host via the bastion host: `../../build/cog -v ssh -u ubuntu cog_target`.

Here's the payoff! You can now use `cog` to SSH into a target host via a bastion host in one step, with no managed personal SSH certificates. Your output will look something like this:

```
1_docker_no_gcp $ ../../build/cog -v ssh -u ubuntu cog_target
INFO[0000] buildVersion: development buildDate: 2020-09-15 16:57:32 UTC buildHash: development buildOS: darwin
INFO[0000] Using inventory: /Users/pemerson/.config/cog/inventory.yaml
INFO[0000] Reading /Users/pemerson/.config/cog/inventory.yaml
INFO[0000] Getting cert group and SSH group for cog_target
INFO[0000] Cert group: ssh, SSH group: global
INFO[0000] Looking up bastion host for environment global.ssh
INFO[0000] bastion user: demouser
INFO[0000] ssh user: ubuntu
INFO[0000] Using existing SSH agent
INFO[0000] Generating a new SSH key
INFO[0000] SSH Mount Point: ssh
INFO[0000] SSH Role: demouser
INFO[0000] Using vault server http://127.0.0.1:8200
INFO[0000] No IAP authorization detected
INFO[0000] Found token in ~/.vault-token
INFO[0000] Using existing credentials
INFO[0000] Adding certificate to SSH agent
INFO[0000] Command: /usr/bin/ssh -J demouser@cog_bastion ubuntu@cog_target
Welcome to Ubuntu 20.04 LTS (GNU/Linux 4.19.76-linuxkit x86_64)
...
$
```

### 8_run_ssh.sh

This script is simply a single SSH command: `ssh ubuntu@cog_target`. This is what most people would want to run most of the time. Due to the configuration of SSH in `~/.ssh/config`, there's a lot going on underneath the hood, but it's completely transparent to the end user.

Once you're SSHed in to the target host, try typing `hostname` and `whoami` to see the current host and user.

## Cleanup

To clean up, simply run the `./1_clear.sh` script.
