#!/bin/bash

# WeDeploy CLI Tool test script

set -euo pipefail
IFS=$'\n\t'

cd $(dirname $0)/..

skipIntegrationTests=false

function helpmenu() {
  echo "WeDeploy CLI Tool test script:

Use ./test.sh [flags]

Flags:
--skip-integration-tests: skip integration tests (not recommended)"
  exit 1
}

if [ "$#" -ne 0 ] && [ $1 == "help" ]; then
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
    esac
    shift
done

function checkWorkingDir() {
  if [ $(git status --short | wc -l) -gt 0 ]; then
    echo "You have uncommited changes."
    git status --short
  fi
}

function runTests() {
  echo "Checking for unchecked errors."
  errcheck $(go list ./... | grep -v /vendor/)
  echo "Linting code."
  test -z "$(golint ./... | grep -v "^vendor" | tee /dev/stderr)"
  echo "Examining source code against code defect."
  go vet $(go list ./... | grep -v /vendor/)
  echo "Checking if code can be simplified or can be improved."
  megacheck ./...
  echo "Running tests (may take a while)."

  # skip integration testing with -race because pseudoterm has some data race condition at this time
  go test $(go list ./... | grep -v /vendor/ | grep -v /integration$) -race
  if [[ $skipIntegrationTests != true ]] ; then
    go test github.com/wedeploy/cli/integration
  fi
}

function run() {
  CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
  LAST_TAG="$(git describe HEAD --tags --abbrev=0 2> /dev/null)" || true

  checkWorkingDir
  runTests
  echo All tests and checks necessary for release passed.
}

run
