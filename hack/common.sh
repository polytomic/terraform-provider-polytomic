#!/bin/bash

LOCAL_VERSION="99.0.0"
PLUGIN_DIR="$HOME/.terraform.d/plugins"
MODULE_PATH="terraform.local/local/polytomic"
TEMPLATE_DIR="$HOME/polytomic-terraform-test"


# Get operating system and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')

LOCAL_PROVIDER_PATH="$PLUGIN_DIR/$MODULE_PATH/$LOCAL_VERSION/${OS}_${ARCH}"
