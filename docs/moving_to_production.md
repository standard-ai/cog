# Moving `cog` to production

Both of the [examples](../examples/) implement all of these steps, so it may be useful to look at them in addition to the documentation below.

## Building the `cog` binary

There is a [build script template](../scripts/build.sh.template) that you should copy to `build.sh` and edit. There are eight configurations to make, detailed below.

| Variable | Description |
| :------- | :---------- |
| `default_vaultAddress` | This is the external domain of Vault, prefixed with `https://`. Defined by `vault_external_domain` in [terraform.tfvars](../terraform/1_gcp_infrastructure/terraform.tfvars.template) |
| `default_vaultProxyHost` | This is the internal domain of Vault, without any prefix. Defined by `vault_internal_domain` in [terraform.tfvars](../terraform/1_gcp_infrastructure/terraform.tfvars.template) |
| `default_vaultIAPServiceAccount` | The service account to use when establishing IAP tunnel. Defined as `service_account_id`@`gcp_project_id`.iam.gserviceaccount.com in [terraform.tfvars](../terraform/1_gcp_infrastructure/terraform.tfvars.template) |
| `default_vaultIAPClientID` | Defined by `oauth2_client_id` in [terraform.tfvars](../terraform/1_gcp_infrastructure/terraform.tfvars.template) |
| `default_gcsBucket` | Defined by `gcp_cog_storage_bucket_name` in [terraform.tfvars](../terraform/1_gcp_infrastructure/terraform.tfvars.template) |
| `default_gcsFilename` | `inventory.yaml` |
| `default_binaryGCSBucket` | When defined, `cog` will use this GCS bucket to do binary upgrades via `cog upgrade` |
| `default_binaryGCSPath` | When defined, `cog` will look in this GCS folder to do binary upgrades via `cog upgrade` |

Once these are defined, running `make` in the top level directory of the repository will build a `cog` binary in the `src/build` directory.

## Creating inventory

The inventory file usually resides in `~/.config/cog/inventory.yaml`. It is recommended to put this inventory file in a GCS bucket for easy distribution with `cog`. Here's an example inventory file:

```
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
aliases:
  cog_target:
  - mytarget
```

First in the YAML is a map of bastion hosts to target hosts. Each array consists of:

* A list of bastions for the group of hosts

* A list of globs

If your inventory is really dynamic, you probably don't want to pull down new inventory every time there's a change. The globs allow you to pattern match so that `cog` knows
which certificate authority to use.

* A list of hosts

Given the glob configuration above, a list of hosts is not strictly necessary. However, `cog` has a nice feature: if you type `cog ssh foo`, it will return a list of hosts that match `foo` if that list is longer than 1. If the length of the list is equal to 1, it will SSH to that host.

* The SSH Certificate Authority

If you decide to configure Vault with multiple Certificate Authorities instead of the one as in the examples, this configures which Certificate Authority to use for these target hosts.

* The SSH group

Since you can map multiple bastions to groups of hosts, `cog` will ping the bastions for you in order to pick out the "best" one. The end result in the `cog.yaml` configuration file is a list of "best" bastions for each SSH group.

Also in the YAML is a map of hosts to a list of aliases. This allows you to create shortcut names for your hosts, which is useful if you have multiple hostname conventions across your organization.

## Configuring local SSH to enable cog

In order to use SSH with `cog` behind the scenes as a ProxyCommand, your local SSH will need to be configured. It is recommended to put this inventory file in a GCS bucket for easy distribution with `cog`. `cog` will inject it into your `~/.ssh/config` file without ruining any existing configuration. Here's an example:

```
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
```

There are two sections of the configuration. One is for bastions, for which we don't want to SSH through a bastion to reach (otherwise we'd have an endless loop), and one for target hosts where we do want to SSH through a bastion server. This section is configured with a ProxyCommand to call `cog` behind the scenes.

Of particular note is the `UserKnownHostsFile`, which is populated with keys from the Certificate Authority to verify hosts that you SSH into.

## Creating a known_hosts file for client-side host verification

In order for clients to verify the hosts it is connecting to, `cog` uses a `~/.config/cog/known_hosts` file, and if configured will pull that file down from GCS on `cog init`.

The file looks like this:

```
@cert-authority <GLOB> <PUBLIC KEY>
```
The public key can be retrieved using vault:

```
vault read -field=public_key ssh-hostkey/config/ca
```

You may need to establish a tunnel and add `--tls-skip-verify`.

If `cog` is configured, you can retrieve it by proxying to Vault:

```
cog vault-proxy run vault -- read -field=public_key ssh-hostkey/config/ca
```

## Trusted user CA keys

The trusted user CA keys get installed on each host. It can be retrieved from Vault:

```
cog vault-proxy run vault -- read -field=public_key ssh/config/ca
```

## Host key signing

Host keys need signing by the Certificate Authority so that they can be verified by the `cog` `known_hosts` file.

```
cog vault-proxy run vault -- write -field=signed_key ssh-hostkey/sign/ssh-hostkey cert_type=host public_key=@path/to/key.pub > path/to/key-cert.pub
```

## Configuring SSH on your hosts

SSH must be configured on each host to allow signed certificates to work properly.

Here's the recommended additions to a typical `/etc/ssh/sshd_config`:

```
HostKey /etc/ssh/ssh_host_ed25519_key
HostCertificate /etc/ssh/ssh_host_ed25519_key-cert.pub
TrustedUserCAKeys /etc/ssh/trusted-user-ca-keys.pem

Match User ubuntu
    AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u
```

For each host, then, you would upload or add:

* An `sshd_config` file
* private host key (may exist already)
* public host key (may exist already)
* signed public host key
* a trusted user CA key
* Any desired files for authorized principals

With a role named `superuser` in /etc/ssh/auth_principals/ubuntu, a user granted that role in Hashicorp Vault would be able to SSH in to the `ubuntu` user on the host.
