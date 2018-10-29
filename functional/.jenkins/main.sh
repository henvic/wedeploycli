#!/bin/bash

set -o errexit
set -o xtrace

CURRENT_DIR=$(pwd)

main() {
  case $1 in

  --log-into-dockerhub)
    log_into_dockerhub
    ;;

  --pull-infrastructure-images)
    pull_infrastructure_images
    ;;

  --setup-machine)
    setup_machine
    ;;

  --shutdown-infrastructure)
    shutdown_infrastructure
    ;;

  --start-infrastructure)
    start_infrastructure
    ;;

  *)
    echo "Error:
  Unknown command.
  Usage: main.sh [--create-test-user|--log-into-dockerhub|--pull-infrastructure-images|
    --setup-machine|--shutdown-infrastructure|--start-infrastructure]

  Aborting."
    exit 1
    ;;
  esac
}

setup_machine() {
  rm -rf "$CURRENT_DIR/.runner/ci-infrastructure"

  echo "Fetching exploded infra"

  mkdir -p "$CURRENT_DIR/.runner"

  cd "$CURRENT_DIR/.runner"

  git clone https://c2b67bf29e8283e809dbf99360074d7f8881d77e@github.com/wedeploy/ci-infrastructure.git

  chmod +x "$CURRENT_DIR/.runner/ci-infrastructure/runner/exploded-infra-runner.sh"

  dig +trace wedeploy.domains

  curl https://cdn.wedeploy.com/cli/latest/wedeploy.sh -fsSL | sudo bash
}

start_infrastructure() {
  local BUILD_TAG=staging
  local WEDEPLOY_ENVIRONMENT=wd-paas-test-us-east-1
  bash "$CURRENT_DIR/.runner/ci-infrastructure/runner/exploded-infra-runner.sh" --run $BUILD_TAG $WEDEPLOY_ENVIRONMENT
}

shutdown_infrastructure() {
  bash "$CURRENT_DIR/.runner/ci-infrastructure/runner/exploded-infra-runner.sh"  --shutdown
}

pull_infrastructure_images() {
  sh "$CURRENT_DIR/.runner/ci-infrastructure/runner/exploded-infra-runner.sh" --pull-images
}

log_into_dockerhub() {
  echo "INFO:
  logging into docker hub.
  "
  bash -ci 'docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD'
}

main "$@"
