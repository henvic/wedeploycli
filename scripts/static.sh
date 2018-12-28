#!/bin/bash

# WeDeploy CLI Tool static analysis scripts

set -euo pipefail
IFS=$'\n\t'

cd $(dirname $0)/..

echo "Checking for unchecked errors."
errcheck $(go list ./...)
echo "Linting code."
test -z "$(golint ./... | grep -v "^vendor" | tee /dev/stderr)"
echo "Examining source code against code defect."
go vet -shadow $(go list ./...)
echo "Checking if code can be simplified or can be improved."
megacheck ./...
