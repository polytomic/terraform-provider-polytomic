#!/bin/bash
set -e

source $(dirname $0)/common.sh


mkdir -p $TEMPLATE_DIR

cat <<EOF > $TEMPLATE_DIR/main.tf
terraform {
  required_providers {
    polytomic = {
      source  = "$MODULE_PATH"
      version = "$LOCAL_VERSION"
    }
  }
}

provider "polytomic" {
  deployment_url     = "app.polytomic-local.com:8443"
  deployment_api_key = "secret-key"
}
EOF

rm -rf $TEMPLATE_DIR/.terraform
rm -f $TEMPLATE_DIR/.terraform.lock.hcl

pushd $TEMPLATE_DIR
terraform init
popd

echo "Created templated terraform project in $TEMPLATE_DIR"