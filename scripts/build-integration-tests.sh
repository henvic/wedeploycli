#!/bin/bash

# Liferay CLI integration tests build tool

set -euo pipefail
IFS=$'\n\t'

PLATFORMS=(darwin linux windows)

cd $(dirname $0)/../integration
INTEGRATION_TESTS_DIR=$PWD

echo "Building integration test suites for multiple platforms:"
WEDEPLOY_CLI_INTEGRATION_TESTS_PATH=${WEDEPLOY_CLI_INTEGRATION_TESTS_PATH:-"."}

function checkCONT() {
  if [[ $CONT != "y" && $CONT != "yes" ]]; then
    exit 1
  fi
}

function removeAndCopyMocks() {
  (which trash >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    (trash mocks 2>/dev/null &&
      >&2 echo -e "\x1B[101m\x1B[1mWarning: mocks directory placed on trash and recreated.\x1B[0m") || true
  else
    read -p "Continue removing mocks to copy them back? [no]: " CONT < /dev/tty;
    checkCONT
    rm -r mocks 2>/dev/null || true
  fi
  cd - >> /dev/null
  cp -r mocks $WEDEPLOY_CLI_INTEGRATION_TESTS_PATH/mocks
}

for platform in ${PLATFORMS[@]}; do
  out="wedeploy-cli-integration-${platform}.test"

  if [[ $platform == "windows" ]]; then
  	out=${out}.exe
  fi

  echo -ne "${platform}...\t"
  env GOOS=${platform} go test -c -tags=nocompile -o ${WEDEPLOY_CLI_INTEGRATION_TESTS_PATH}/$out
  echo $out
done

cd $WEDEPLOY_CLI_INTEGRATION_TESTS_PATH

if [ ! $WEDEPLOY_CLI_INTEGRATION_TESTS_PATH -ef $INTEGRATION_TESTS_DIR ]; then
  removeAndCopyMocks
fi

echo
echo "Integration test suites and its related mocks are saved in:"
echo $(pwd)

if [ $WEDEPLOY_CLI_INTEGRATION_TESTS_PATH -ef $INTEGRATION_TESTS_DIR ]; then
  echo
  echo "Set the \$WEDEPLOY_CLI_INTEGRATION_TESTS_PATH environment variable to save somewhere else."
fi
