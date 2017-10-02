#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

dep ensure
echo "Regenerating legal/licenses.go"
go generate github.com/wedeploy/cli/legal

missing=$(vendorlicenses -missing)
if [[ $missing ]]; then
    echo "Following dependencies are missing license files:"
    echo "$missing"
else
    echo "Found licenses for all dependencies."
fi

