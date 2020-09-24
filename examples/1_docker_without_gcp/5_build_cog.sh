#!/bin/bash

export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=$(cat vault_initialization.log | grep 'Initial Root Token' | cut -d : -f 2 | sed 's/ //g')
export USERNAME=demouser

mkdir -p ~/.config/cog

# Create ~/.config/cog/cog.yaml
cat > ~/.config/cog/cog.yaml <<EOF
bastion_map:
  global.ssh: cog_bastion
bastion_user: ${USERNAME}
ssh_user: ubuntu
vault_address: http://127.0.0.1:8200
EOF

# Create ~/.config/cog/inventory.yaml
cat > ~/.config/cog/inventory.yaml <<EOF
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
EOF

# Add CA hostkey to ~/.config/cog/known_hosts
echo "@cert-authority cog_target $(vault read -tls-skip-verify -field=public_key ssh-hostkey/config/ca)" > ~/.config/cog/known_hosts
echo "@cert-authority [127.0.0.1]:2222 $(vault read -tls-skip-verify -field=public_key ssh-hostkey/config/ca)" >> ~/.config/cog/known_hosts

# Add configuration to ~/.ssh/config
grep 'DEMO COG CONFIG' ~/.ssh/config
exit=$?
if [ $exit -eq 0 ] ; then
  echo "Not manipulating ~/.ssh/config"
else
  cd ../../
  cog_path="$(pwd)/build/cog"
  cd -
  mkdir -p ~/.ssh
  cat >> ~/.ssh/config <<EOF

### BEGIN DEMO COG CONFIG
Host cog_bastion
    User ${USERNAME}
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
    ProxyCommand ${cog_path} sshproxy -W %h:%p
    UserKnownHostsFile ~/.config/cog/known_hosts

### END DEMO COG CONFIG

EOF
fi

# Build cog binary
cd ../..
make
