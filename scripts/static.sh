#!/bin/bash

# WeDeploy CLI Tool static analysis scripts

set -euo pipefail
IFS=$'\n\t'

cd $(dirname $0)/..

echo "Checking for unchecked errors."
errcheck $(go list ./...)
echo "Linting code."
test -z "$(golint `go list ./...` | tee /dev/stderr)"
echo "Examining source code against code defect."
go vet -shadow $(go list ./...)
echo "Running staticcheck toolset https://staticcheck.io"
staticcheck ./...
echo "Checking if code contains security issues."
# gosec -quiet . # should be ./..., but must fix first something...