#!/bin/bash

# Liferay Cloud Platform CLI Tool installer publishing script

set -euo pipefail
IFS=$'\n\t'

cd $(dirname $0)/..

config=""
INSTALLER_RELEASE_CHANNEL="installer"
INSTALLER_RELEASE_VERSION="installer-0.2"

function helpmenu() {
  echo "Liferay Cloud Platform CLI Tool installer publishing script:

1) check if all changes are commited
2) create and push a release tag
3) build and push a new release to equinox

Use ./release.sh [flags]

Flags:
--config: release configuration file"
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

function checkCONT() {
  if [[ $CONT != "y" && $CONT != "yes" ]]; then
    exit 1
  fi
}

function checkWorkingDir() {
  if [ $(git status --short | wc -l) -gt 0 ]; then
    echo "You have uncommited changes."
    git status --short

    echo -e "\x1B[101m\x1B[1mBy continuing you might release a build with incorrect content.\x1B[0m"
    read -p "Continue? [no]: " CONT < /dev/tty;
    checkCONT
  fi
}

function publish() {
  cd update/internal/installer
  equinox release \
  --version=$INSTALLER_RELEASE_VERSION \
  --channel=$INSTALLER_RELEASE_CHANNEL \
  --config=$config \
  -- \
  -ldflags="-X 'main.Version=$INSTALLER_RELEASE_VERSION' \
  -X 'main.Build=$BUILD_COMMIT' \
  -X 'main.BuildTime=$BUILD_TIME'" \
  -gcflags=-trimpath=$GOPATH \
  -asmflags=-trimpath=$GOPATH \
  github.com/wedeploy/cli/update/internal/installer
  cd ../..
}

function prerelease() {
  checkWorkingDir
  ./scripts/static.sh
  echo All tests and checks necessary for release passed.
}

function release() {
  BUILD_COMMIT=$(git rev-list -n 1 HEAD)
  BUILD_TIME=$(date -u)

  echo "build commit $BUILD_COMMIT at $BUILD_TIME"

  publish
}

function run() {
  prerelease
  echo
  release
}

run
