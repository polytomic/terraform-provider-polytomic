#!/bin/bash
# This script builds the polytomic terraform provider
# and installs it in the local terraform plugin directory.
set -e

source $(dirname $0)/common.sh

echo "Building provider..."
go install

echo "Copying provider to local terraform plugin directory..."
cp $GOPATH/bin/terraform-provider-polytomic $LOCAL_PROVIDER_PATH
