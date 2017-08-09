#!/bin/bash

# WeDeploy CLI Tool publishing script

set -euo pipefail
IFS=$'\n\t'

cd `dirname $0`/..

skipIntegrationTests=false
config=""

function helpmenu() {
  echo "WeDeploy CLI Tool publishing script:

1) check if all changes are commited
2) run tests on a local drone.io instance
3) create and push a release tag
5) build and push a new release to equinox

Use ./release.sh [flags]

Flags:
--config: release configuration file
--skip-integration-tests: skip integration tests (not recommended)"
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
        --skip-integration-tests)
            skipIntegrationTests=true
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

function testHuman() {
  if [ ! -t 1 ] ; then
    >&2 echo "Run this script from terminal."
    exit 1
  fi
}

function checkWeDeployImageTag() {
  cat defaults/defaults.go | grep -q "WeDeployImageTag = \"latest\"" && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    >&2 echo -e "\x1B[101m\x1B[1mWarning: you MUST NOT use docker image tag \"latest\" for releases.\x1B[0m"
    read -p "Continue? [no]: " CONT < /dev/tty;
    checkCONT
  fi
}

function checkWorkingDir() {
  if [ `git status --short | wc -l` -gt 0 ]; then
    echo "You have uncommited changes."
    git status --short

    echo -e "\x1B[101m\x1B[1mBy continuing you might release a build with incorrect content.\x1B[0m"
    read -p "Continue? [no]: " CONT < /dev/tty;
    checkCONT
  fi
}

function runTests() {
  echo "Checking for unchecked errors."
  errcheck $(go list ./... | grep -v /vendor/)
  echo "Linting code."
  test -z "$(golint ./... | grep -v "^vendor" | tee /dev/stderr)"
  echo "Examining source code against code defect."
  go vet $(go list ./... | grep -v /vendor/)
  echo "Running tests (may take a while)."

  # skip integration testing with -race because pseudoterm has some data race condition at this time
  go test $(go list ./... | grep -v /vendor/ | grep -v /integration$) -race
  if [[ $skipIntegrationTests != true ]] ; then
    go test $(go list ./... | grep -v /vendor/)
  fi
}

function runTestsOnDrone() {
  (which drone >> /dev/null) && ec=$? || ec=$?

  if [ ! $ec -eq 0 ] ; then
    >&2 echo "Error trying to run tests on drone."
    read -p "drone not found on your system: Release anyway? [no]: " CONT < /dev/tty
    checkCONT
    return
  fi

  echo "Running tests isolated on drone (docker based CI) locally."
  drone exec && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    return
  fi

  read -p "Tests on drone failed: Release anyway? [no]: " CONT < /dev/tty
  checkCONT
}

function publish() {
  equinox release \
  --version=$NEW_RELEASE_VERSION \
  --channel=$RELEASE_CHANNEL \
  --config=$config \
  -- \
  -ldflags="-X 'github.com/wedeploy/cli/defaults.Version=$NEW_RELEASE_VERSION' \
  -X 'github.com/wedeploy/cli/defaults.Build=$BUILD_COMMIT' \
  -X 'github.com/wedeploy/cli/defaults.BuildTime=$BUILD_TIME'" \
  -gcflags=-trimpath=$GOPATH \
  -asmflags=-trimpath=$GOPATH \
  github.com/wedeploy/cli
}

function prerelease() {
  testHuman
  checkWeDeployImageTag
  checkWorkingDir
  runTests
  echo All tests and checks necessary for release passed.
}

function release() {
  read -p "Release channel [unstable]: " RELEASE_CHANNEL < /dev/tty;
  RELEASE_CHANNEL=${RELEASE_CHANNEL:-"unstable"}

  BUILD_COMMIT=`git rev-list -n 1 HEAD`
  BUILD_TIME=`date -u`

  echo "build commit $BUILD_COMMIT at $BUILD_TIME"

  runTestsOnDrone
  publish
}

function checkTag() {
  CURRENT_TAG=`git describe --exact-match HEAD` || true

  if [[ $CURRENT_TAG == "" ]] ; then
    echo "Maybe you want to pull changes"
    exit 1
  fi

  read -p "Confirm release version (tag): " NEW_RELEASE_VERSION < /dev/tty;
  # normalize by removing leading v (i.e., v0.0.1)
  NEW_RELEASE_VERSION=`echo $NEW_RELEASE_VERSION | sed 's/^v//'`

  if [[ $CURRENT_TAG != "v$NEW_RELEASE_VERSION" ]] ; then
    echo "Current tag is $CURRENT_TAG, but you tried to release v$NEW_RELEASE_VERSION"
    exit 1
  fi
}

function run() {
  checkTag
  prerelease
  echo
  release
}

run
