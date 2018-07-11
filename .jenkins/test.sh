#!/bin/bash

set -o errexit
set -o xtrace

main() {
  cd functional/tests
  TEAMUSER_EMAIL="qa.team.user@wedeploy.com" \
  TEAMUSER_PW="L6P&ZExVXydC" \
  TESTER_EMAIL="cli-tester@test.com" \
  TESTER_PW="test" \
  REMOTE=wedeploy.xyz \
  ./main.exp
}

main "$@"
