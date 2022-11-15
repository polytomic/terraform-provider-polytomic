#!/bin/bash
# This script sets up a local development environment for the polytomic terraform provider.
set -e

source $(dirname $0)/common.sh

echo "Writing .terraformrc config... "
cat <<EOF > ~/.terraformrc
provider_installation {
  filesystem_mirror {
    path    = "$PLUGIN_DIR"
  }
  direct {
    exclude = ["terraform.local/*/*"]
  }
}
EOF
echo ".terraformrc config written"

echo "Creating local provider directory ${LOCAL_PROVIDER_PATH} ... "
mkdir -p $LOCAL_PROVIDER_PATH

echo "Clear out any existing local provider binaries... "
rm -f $LOCAL_PROVIDER_PATH/*

cat <<EOF

Ensure the following block is set in your terraform configuration:

terraform {
  required_providers {
    polytomic = {
      source  = "$MODULE_PATH"
      version = "$LOCAL_VERSION"
    }
  }
}
EOF