#!/bin/bash

# Liferay Cloud Platform CLI tool build and run script (with race detector)

set -euo pipefail
IFS=$'\n\t'

cd $GOPATH/src/github.com/wedeploy/cli
go install -race
cd ~-
$GOPATH/bin/cli $@ && ec=$? || ec=$?
if [ ! $ec -eq 0 ] ; then
  echo "exit status $ec"
fi
