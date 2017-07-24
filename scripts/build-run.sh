#!/bin/bash

# WeDeploy CLI tool build and run script (with race detector)

set -euo pipefail
IFS=$'\n\t'

cd $GOPATH/src/github.com/wedeploy/cli
go install -race
cd -
$GOPATH/bin/cli $@
