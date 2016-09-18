#!/bin/bash

# WeDeploy CLI tool publishing script

set -euo pipefail
IFS=$'\n\t'

config=${1-}

if [ -z $config ] || [ $config == "help" ] || [ $config == "--help" ]; then
  echo "WeDeploy CLI Tool publishing script:

1) check if all changes are commited
2) run tests on a local drone.io instance
3) create and push a release tag
5) build and push a new release to equinox

Check Semantic Versioning rules on semver.org

Use ./release.sh <equinox config> to release
Use ./release.sh --pre-release to run release tests only"
  exit 1
fi

function checkCONT() {
  if [[ $CONT != "y" && $CONT != "yes" ]]; then
    exit 1
  fi
}

function testHuman() {
  if [ ! -t 1 ] ; then
    >&2 echo "Run this script from terminal."
    exit 1
  fi
}

function checkBranch() {
  if [ $CURRENT_BRANCH != "master" ]; then
    read -p "Draft a new release from non-default $CURRENT_BRANCH branch [no]: " CONT < /dev/tty;
    checkCONT
  fi
}

function checkWorkingDir() {
  if [ `git status --short | wc -l` -gt 0 ]; then
    echo "You have uncommited changes."
    git status --short

    if [ $config != "--pre-release" ]; then
      echo -e "\x1B[101m\x1B[1mBy continuing you might generate a release tag with incorrect content.\x1B[0m"
      read -p "Continue? [no]: " CONT < /dev/tty;
      checkCONT
    else
      echo -e "\x1B[101m\x1B[1mPlease commit your changes before generating a new release tag.\x1B[0m"
    fi
  fi
}

function checkPublishedTag() {
  echo "Verifying if tag v$NEW_RELEASE_VERSION is already published on origin remote."
  check_published_tag=`git ls-remote origin refs/tags/v$NEW_RELEASE_VERSION | wc -l`
  if [ "$check_published_tag" -gt 0 ]; then
    >&2 echo "git tag v$NEW_RELEASE_VERSION already published. Not forcing update."
    exit 1
  fi
  echo
}

function checkUnusedTag() {
  OVERWRITE_TAG=false
  check_tags_free=`git tag --list "v$NEW_RELEASE_VERSION" | wc -l`
  if [ "$check_tags_free" -gt 0 ]; then
    read -p "git tag exists locally. Overwrite? [yes]: " CONT < /dev/tty;
    CONT=${CONT:-"yes"}
    checkCONT
    OVERWRITE_TAG=true
  fi
}

function runTests() {
  echo "Checking for unchecked errors."
  errcheck $(go list ./... | grep -v /vendor/)
  echo "Linting code."
  test -z "$(golint ./... | grep -v "^vendor" | tee /dev/stderr)"
  echo "Examining source code against code defect."
  go vet $(go list ./... | grep -v /vendor/)
  echo "Running tests (may take a while)."
  go test $(go list ./... | grep -v /vendor/)
}

function runTestsOnDrone() {
  echo "Running tests isolated on drone (docker based CI) locally."
  drone exec && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    return
  fi

  read -p "Tests on drone failed: Release anyway? [no]: " CONT < /dev/tty
  checkCONT
}

function setEditor() {
  editor="${EDITOR:editor}"
  editor=`git config core.editor`
  editor="${editor:-vi}"
}

function tag() {
  echo "Waiting for a ChangeLog / release summary."

  if [ $OVERWRITE_TAG == true ] ; then
    echo "Overwriting unpublished tag v$NEW_RELEASE_VERSION."
    git tag -s "v$NEW_RELEASE_VERSION" --force
    return
  fi

  # if tag doesn't exists we want to create a release message
  if [ $LAST_TAG == "" ] ; then
    >&2 echo "Can't find previous release tag."
    LAST_TAG="HEAD"
  fi

  setEditor

  (
    echo "# Lines starting with '#' will be ignored."
    echo "Release v$NEW_RELEASE_VERSION"
    echo ""
    echo "Changes:"
    git log $LAST_TAG..HEAD --pretty="format:%h %s" --abbrev=10 || true
    echo ""
    echo ""
    echo "Build commit:"
    echo $BUILD_COMMIT
    echo ""
  ) > .git/TAG_EDITMSG

  bash -c "$editor .git/TAG_EDITMSG"
  git tag -s "v$NEW_RELEASE_VERSION" -F .git/TAG_EDITMSG
}

function publish() {
  tag
  git checkout "v$NEW_RELEASE_VERSION"

  equinox release \
  --version=$NEW_RELEASE_VERSION \
  --channel=$RELEASE_CHANNEL \
  --config=$config \
  -- \
  -ldflags="-X github.com/wedeploy/cli/defaults.Version=$NEW_RELEASE_VERSION \
  -X github.com/wedeploy/cli/defaults.Build=$BUILD_COMMIT" \
  -gcflags=-trimpath=$GOPATH \
  -asmflags=-trimpath=$GOPATH \
  github.com/wedeploy/cli

  git push origin "v$NEW_RELEASE_VERSION"
  git checkout $CURRENT_BRANCH
}

function prerelease() {
  testHuman

  if [ $config != "--pre-release" ]; then
    checkBranch
  fi

  checkWorkingDir
  runTests
  echo All tests and checks necessary for release passed.
}

function release() {
  echo Release announcements should use semantic versioning.

  if [ $LAST_TAG != "" ] ; then
    echo Last version seems to be $LAST_TAG
  fi

  read -p "Release channel [stable]: " RELEASE_CHANNEL < /dev/tty;
  RELEASE_CHANNEL=${RELEASE_CHANNEL:-"stable"}

  read -p "New version: " NEW_RELEASE_VERSION < /dev/tty;
  # normalize by removing leading v (i.e., v0.0.1)
  NEW_RELEASE_VERSION=`echo $NEW_RELEASE_VERSION | sed 's/^v//'`
  BUILD_COMMIT=`git rev-list -n 1 HEAD`

  echo "build commit $BUILD_COMMIT"

  runTestsOnDrone
  checkPublishedTag
  checkUnusedTag
  publish
}

function run() {
  CURRENT_BRANCH=`git rev-parse --abbrev-ref HEAD`
  LAST_TAG="$(git describe HEAD --tags --abbrev=0 2> /dev/null)" || true

  if [ $config == "--pre-release" ]; then
    prerelease
  else
    prerelease
    echo
    release
  fi
}

run
