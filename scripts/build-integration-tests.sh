#!/bin/bash

# WeDeploy CLI integration tests build tool

set -euo pipefail
IFS=$'\n\t'

PLATFORMS=(darwin linux windows)

cd `dirname $0`/../integration
INTEGRATION_TESTS_DIR=$PWD

echo "Building integration test suites for multiple platforms:"
WEDEPLOY_CLI_INTEGRATION_TESTS_PATH=${WEDEPLOY_CLI_INTEGRATION_TESTS_PATH:-"."}

for platform in ${PLATFORMS[@]}; do
  out="wedeploy-cli-integration-${platform}.test"

  if [[ $platform == "windows" ]]; then
  	out=${out}.exe
  fi

  echo -ne "${platform}...\t"
  env GOOS=${platform} go test -c -tags=integration -o ${WEDEPLOY_CLI_INTEGRATION_TESTS_PATH}/$out
  echo $out
done

cd $WEDEPLOY_CLI_INTEGRATION_TESTS_PATH
echo
echo "Integration test suites are saved in:"
echo `pwd`

if [ ! $WEDEPLOY_CLI_INTEGRATION_TESTS_PATH -ef $INTEGRATION_TESTS_DIR ]; then
  echo
  echo "Set the \$WEDEPLOY_CLI_INTEGRATION_TESTS_PATH environment variable to save somewhere else."
fi
