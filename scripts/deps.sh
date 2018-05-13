#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

dep ensure
$(dirname "$0")/legal.sh
