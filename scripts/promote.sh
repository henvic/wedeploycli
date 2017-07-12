#!/bin/bash

# WeDeploy CLI tool version promoting script

set -euo pipefail
IFS=$'\n\t'

config=""

function helpmenu() {
  echo "WeDeploy CLI Tool version promoting script:

Use ./promote.sh --config <file> to promote version"
  exit 1
}

if [ "$#" -eq 0 ] || [ $1 == "help" ]; then
  helpmenu
fi

while [ ! $# -eq 0 ]
do
    case "$1" in
        --help | -h)
            helpmenu
            ;;
        --config)
            config=${2-}
            shift
            break
            ;;
    esac
    shift
done

read -p "Version: " NEW_RELEASE_VERSION < /dev/tty;
NEW_RELEASE_VERSION=$(echo $NEW_RELEASE_VERSION | sed 's/^v//')

read -p "Promote version to release channel [stable]: " RELEASE_CHANNEL < /dev/tty;
RELEASE_CHANNEL=${RELEASE_CHANNEL:-"stable"}

equinox publish --channel $RELEASE_CHANNEL --config $config --release $NEW_RELEASE_VERSION
