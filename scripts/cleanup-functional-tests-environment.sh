#!/bin/bash

# WeDeploy CLI Tool cleanup functional test environment script

set -euo pipefail
IFS=$'\n\t'

cd $(dirname $0)/../functional

./cleanup-environment.exp
