#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

echo "Regenerating legal/licenses.go"
go generate github.com/henvic/wedeploycli/legal

missing=$(vendorlicenses -missing)
if [[ $missing ]]; then
    echo "Following dependencies are missing license files:"
    echo "$missing"
else
    echo "Found licenses for all dependencies."
fi
