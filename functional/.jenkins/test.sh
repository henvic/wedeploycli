#!/bin/bash

set -o errexit
set -o xtrace

main() {
  cd ../tests
  ./main.exp
}

main "$@"
