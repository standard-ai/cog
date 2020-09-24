# Demoing cog with GCP infrastructure

## Overview

This demo uses vault and docker containers (via docker-compose) to create a fully running
example of `cog`.

## Prerequisites

* Successful terraforming of the infrastructure in the [terraform](../../terraform) directory.
* [HashiCorp Terraform 0.13](https://www.terraform.io/downloads.html)
* [HashiCorp Vault](https://www.vaultproject.io/downloads)
* [Docker](https://www.docker.com/products/docker-desktop)
* [Golang](https://golang.org/dl/)

##

## TL;DR

```
./build.sh
ssh ubuntu@cog_target
```

## The scripts

These scripts should be run in order. The [build.sh](build.sh) script runs scripts `1_clear.sh` through `3_docker.sh`.

### 1_clear.sh

* Remove vault login
* Remove SSH keys
* Stop and remove containers, networks, images, and volumes
* Remove cog configuration
* Remove cog stanza from ~/.ssh/config

### 2_cog_setup.sh

* Build cog binary
* Create ~/.config/cog/inventory.yaml
* Copy inventory.yaml to GCS
* Copy ssh_config to GCS
* Copy empty known_hosts to GCS
* Initialize cog
* Use `cog vault-proxy run` to create good known_hosts
* Copy known_hosts to GCS
* Re-initialize cog to pull good known_hosts down

### 3_docker.sh

* Use `cog vault-proxy run` to retrieve trusted user CA key from Vault
* Use `cog vault-proxy run` to generate and sign bastion host key
* Use `cog vault-proxy run` to generate and sign target host key
* Build the two host containers
* Run the host containers

### 4_run_cog.sh

This script just runs cog to SSH into the target host via the bastion host: `../../build/cog -v ssh -u ubuntu cog_target`.

Here's the payoff! You can now use `cog` to SSH into a target host via a bastion host in one step, with no managed personal SSH certificates. Your output will look something like this:

```
../../build/cog -v ssh -u ubuntu cog_target
INFO[0000] buildVersion: development buildDate: 2020-09-16 18:32:26 UTC buildHash: development buildOS: darwin
INFO[0000] Using inventory: /Users/pemerson/.config/cog/inventory.yaml
INFO[0000] Reading /Users/pemerson/.config/cog/inventory.yaml
INFO[0000] Getting cert group and SSH group for cog_target
INFO[0000] Cert group: ssh, SSH group: global
INFO[0000] Looking up bastion host for environment global.ssh
INFO[0000] bastion user: pemerson
INFO[0000] ssh user: ubuntu
INFO[0000] Using existing SSH agent
INFO[0000] Generating a new SSH key
INFO[0000] SSH Mount Point: ssh
INFO[0000] SSH Role: pemerson
INFO[0000] Using vault server https://vault-pete-external.nonstandard.ai
INFO[0000] Adding IAP authorization to vault client
INFO[0001] Cached token invalid, attempting vault oidc login
INFO[0007] Adding certificate to SSH agent
INFO[0007] Command: /usr/bin/ssh -J pemerson@cog_bastion ubuntu@cog_target
Welcome to Ubuntu 20.04 LTS (GNU/Linux 4.19.76-linuxkit x86_64)
...
$
```

### 5_run_ssh.sh

This script is simply a single SSH command: `ssh ubuntu@cog_target`. This is what most people would want to run most of the time. Due to the configuration of SSH in `~/.ssh/config`, there's a lot going on underneath the hood, but it's completely transparent to the end user.

Once you're SSHed in to the target host, try typing `hostname` and `whoami` to see the current host and user.

## Cleanup

To clean up, simply run the `./1_clear.sh` script.
