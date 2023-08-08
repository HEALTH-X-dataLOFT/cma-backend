#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 1 ]; then
  cat <<EOF
This script is used for creating a k8s secret with a JWK private key for a
RUN-DSP provider that requires public/private key encryption for passing
through a JWE via the dataspace and a configmap with public key data for
the CMA backend.

NOTE: Do not run this in a production cluster since it will create resources.

Requirements:
- python3
- kubectl
- Configured KUBECONFIG environment variable

After running this script you will of course have to extract the configmap and
secret, and edit them as required before putting in the gitops repository.

EOF
  exit -1
fi
provider_url=$1

if ! command -v kubectl &> /dev/null; then
    echo "'kubectl' tool is missing."
    exit -1
fi

if [ "$KUBECONFIG" = "" ]; then
  echo "Environment variable KUBECONFIG MISSING."
  exit -1
fi

clean_up() {
  popd
  test -d "$tmp_dir" && rm -fr "$tmp_dir"
}

tmp_dir=$( mktemp -d --suffix ".generate_jwk.sh" )
trap "clean_up $tmp_dir" EXIT

echo "Got directory $tmp_dir"
pushd $tmp_dir

cat <<EOF > $tmp_dir/main.py
import os
import subprocess
from jwcrypto import jwk

key_name = "service_app_keys"
key_type = "RSA"
alg = "RSA-OAEP-256"
size = 2048
use = "enc"


def create_keys(key_name):
    key = jwk.JWK.generate(kty=key_type, size=size, kid=key_name, alg=alg)
    with open(f"keys/{key_name}_private.json", "w") as writer:
        writer.write(key.export_private())
    with open(f"keys/{key_name}_public.json", "w") as writer:
        writer.write(key.export_public())

os.makedirs("keys")
create_keys(key_name=key_name)
EOF

env python3 -m venv venv
. $tmp_dir/venv/bin/activate
pip install jwcrypto
python3 main.py

timestamp=$(date +"%s")
public_key=$(cat keys/service_app_keys_public.json | base64 | tr -d "\n")
cat <<EOF > $tmp_dir/public_key_data.json
{
  "${provider_url}": "${public_key}"
}
EOF

kubectl create configmap public-key-data-${timestamp} --from-file=$tmp_dir/public_key_data.json
kubectl create secret generic private-jwk-${timestamp} --from-file=$tmp_dir/keys/service_app_keys_private.json
