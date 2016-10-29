#!/bin/bash

# WeDeploy CLI tool version promoting script

set -euo pipefail
IFS=$'\n\t'

config=${1-}

if [ -z $config ] || [ $config == "help" ] || [ $config == "--help" ]; then
  echo "WeDeploy CLI Tool version promoting script:

Use ./promote.sh <equinox config> to promote version"
  exit 1
fi

read -p "Version: " NEW_RELEASE_VERSION < /dev/tty;
NEW_RELEASE_VERSION=`echo $NEW_RELEASE_VERSION | sed 's/^v//'`

read -p "Promote version to release channel [stable]: " RELEASE_CHANNEL < /dev/tty;
RELEASE_CHANNEL=${RELEASE_CHANNEL:-"stable"}

equinox publish --channel $RELEASE_CHANNEL --config $config --release $VERSION
