#!/bin/bash

# WeDeploy CLI functional tests build tool

set -euo pipefail
IFS=$'\n\t'

PLATFORMS=(darwin linux windows)

cd `dirname $0`/../functional

echo "Building functional test suites for multiple platforms:"
WEDEPLOY_CLI_FUNCTIONAL_TESTS_PATH=${WEDEPLOY_CLI_FUNCTIONAL_TESTS_PATH:-"."}

for platform in ${PLATFORMS[@]}; do
  out="wedeploy-cli-functional-${platform}.test"

  if [[ $platform == "windows" ]]; then
  	out=${out}.exe
  fi

  echo -ne "${platform}...\t"
  env GOOS=${platform} go test -c -o ${WEDEPLOY_CLI_FUNCTIONAL_TESTS_PATH}/$out
  echo $out
done

cd $WEDEPLOY_CLI_FUNCTIONAL_TESTS_PATH
echo
echo "Functional test suites are saved in:"
echo `pwd`

if [ ! -z $WEDEPLOY_CLI_FUNCTIONAL_TESTS_PATH ]; then
  echo
  echo "Set the \$WEDEPLOY_CLI_FUNCTIONAL_TESTS_PATH environment variable to save somewhere else."
fi
