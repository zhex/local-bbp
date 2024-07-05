#!/bin/bash

REPO="zhex/local-bbp"
VERSION="$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep tag_name | cut -d '"' -f 4 | sed 's/^v//')"
BIN_DIR="/usr/local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
  ARCH="amd64"
elif [[ "$ARCH" == "aarch64" ]]; then
  ARCH="arm64"
fi

PACKAGE_NAME="local-bbp_${VERSION}_${OS}_${ARCH}.tar.gz"

RELEASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${PACKAGE_NAME}"
echo $RELEASE_URL

# Download the package
curl -L $RELEASE_URL -o /tmp/$PACKAGE_NAME

# Extract the package
tar -xzvf /tmp/$PACKAGE_NAME -C $BIN_DIR bbp

# Clean up
rm /tmp/$PACKAGE_NAME

echo "Installation complete."
