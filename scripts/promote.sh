#!/bin/bash

# Liferay Cloud Platform CLI tool version promoting script

set -euo pipefail
IFS=$'\n\t'

config=""

function helpmenu() {
  echo "Liferay Cloud Platform CLI Tool version promoting script:

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

echo "Deployment checklist: you must verify that the following commands"
echo "work correctly on macOS, Linux, and Windows:"
echo "lcp login, lcp deploy, lcp shell, lcp list."
echo

read -p "Promote version to release channel [stable]: " RELEASE_CHANNEL < /dev/tty;
RELEASE_CHANNEL=${RELEASE_CHANNEL:-"stable"}

equinox publish --channel $RELEASE_CHANNEL --config $config --release $NEW_RELEASE_VERSION

if [ $RELEASE_CHANNEL == "stable" ] ; then
  echo "Use \"make release-notes-page\" to update https://wedeploy.com/updates/cli/"
  read -p "Don't forget to update the release notes page. Press \"Enter\" when done. " < /dev/tty;
fi
