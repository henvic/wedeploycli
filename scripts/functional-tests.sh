#!/bin/bash

# Liferay CLI Tool functional test script

set -euo pipefail
IFS=$'\n\t'

cd $(dirname $0)/../functional

./main.exp
