#!/bin/bash

apt-get -y update
apt-get -y install ntp docker.io

cat << EOF > /etc/docker/daemon.json
{
   "log-driver" : "syslog",
   "log-opts" : {
      "tag" : "docker: {{.ImageName}}/{{.Name}}/{{.ID}}"
   }
}
EOF

service docker restart

mkdir -p /etc/vault

openssl req -new -newkey rsa:4096 -days 3650 -nodes -x509 \
    -subj "/${openssl_subject}/CN=${vault_internal_domain}" \
    -keyout /etc/vault/${vault_internal_domain}.key  -out /etc/vault/${vault_internal_domain}.cert

cat << EOF > /etc/vault/vault.hcl
ui = true

api_addr = "https://${vault_internal_domain}:8200"

storage "gcs" {
    bucket = "${storage_bucket}"
    ha_enabled =  "true"
}

seal "gcpckms" {
  project     = "${project_id}"
  region      = "${region}"
  key_ring    = "${vault_auto_unseal_key_ring}"
  crypto_key  = "${vault_auto_unseal_crypto_key_name}"
}

listener "tcp" {
	address     = "0.0.0.0:8200"
	tls_cert_file = "/vault/config/${vault_internal_domain}.cert"
	tls_key_file  = "/vault/config/${vault_internal_domain}.key"
}
EOF

docker run \
	-d \
        --restart always \
	--log-driver syslog \
	--cap-add=IPC_LOCK \
	-p 0.0.0.0:8200:8200/tcp \
	-v /dev/log:/dev/log \
	-v /etc/vault:/vault/config \
	--name vault \
	vault:1.4.2 server
