#!/bin/bash

# Liferay Cloud Platform CLI Tool tagging script

set -euo pipefail
IFS=$'\n\t'

cd $(dirname $0)/..

skipIntegrationTests=false
dryrun=false

function helpmenu() {
  echo "Liferay Cloud Platform CLI Tool tagging script:

1) check if all changes are committed
2) run tests on a local drone.io instance
3) create and push a release tag
5) build and push a new release to equinox

Check Semantic Versioning rules on semver.org

Use ./tag.sh [flags]

Flags:
--dry-run: to run release tests only, without tagging a new version
--skip-integration-tests: skip integration tests (not recommended)"
  exit 1
}

if [ "$#" -ne 0 ] && [ $1 == "help" ]; then
  helpmenu
fi

while [ ! $# -eq 0 ]
do
    case "$1" in
        --help | -h)
            helpmenu
            ;;
        --dry-run)
            dryrun=true
            ;;
        --skip-integration-tests)
            skipIntegrationTests=true
            ;;
    esac
    shift
done

function checkCONT() {
  if [[ $CONT != "y" && $CONT != "yes" ]]; then
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
  if [ $(git status --short | wc -l) -gt 0 ]; then
    echo "You have uncommited changes."
    git status --short

    if [ ! $dryrun ]; then
      echo -e "\x1B[101m\x1B[1mBy continuing you might generate a release tag with incorrect content.\x1B[0m"
      read -p "Continue? [no]: " CONT < /dev/tty;
      checkCONT
    else
      echo -e "\x1B[101m\x1B[1mPlease commit your changes before generating a new release tag.\x1B[0m"
    fi
  fi
}

function checkPublishedTag() {
  echo "Verifying if tag v$NEW_TAG_VERSION is already published on origin remote."
  check_published_tag=$(git ls-remote origin refs/tags/v$NEW_TAG_VERSION | wc -l)
  if [ "$check_published_tag" -gt 0 ]; then
    >&2 echo "git tag v$NEW_TAG_VERSION already published. Not forcing update."
    exit 1
  fi
  echo
}

function checkUnusedTag() {
  OVERWRITE_TAG=false
  check_tags_free=$(git tag --list "v$NEW_TAG_VERSION" | wc -l)
  if [ "$check_tags_free" -gt 0 ]; then
    read -p "git tag exists locally. Overwrite? [yes]: " CONT < /dev/tty;
    CONT=${CONT:-"yes"}
    checkCONT
    OVERWRITE_TAG=true
  fi
}

function runTests() {
  ./scripts/static.sh

  echo "Running tests (may take a while)."

  go test $(go list ./... | grep -v /integration$) -race
  if [[ $skipIntegrationTests != true ]] ; then
    go test github.com/henvic/wedeploycli/integration
  fi
}

function runTestsOnDrone() {
  (which drone >> /dev/null) && ec=$? || ec=$?

  if [ ! $ec -eq 0 ] ; then
    >&2 echo "Error trying to run tests on drone."
    read -p "drone not found on your system: Release anyway? [no]: " CONT < /dev/tty
    checkCONT
    return
  fi

  echo "Running tests isolated on drone (docker based CI) locally."
  drone exec && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    return
  fi

  read -p "Tests on drone failed: Tag anyway? [no]: " CONT < /dev/tty
  checkCONT
}

function setEditor() {
  editor="${EDITOR:-vi}"
  gitEditor=$(git config core.editor) || true
  editor="${gitEditor:-$editor}"
}

function tag() {
  (which gpg >> /dev/null) && ec=$? || ec=$?

  if [ ! $ec -eq 0 ] ; then
    >&2 echo "GPG is required to tag signed releases with git. Install it on your system."
    >&2 echo "You might be required to setup a pair of public/private certificates."

    if [ $(uname) == "Darwin" ] ; then
      >&2 echo "Tip: On macOS use https://gpgtools.org instead of \"brew\" to install it."
    fi

    exit 1
  fi

  echo "Waiting for a ChangeLog / release summary."

  if [ $OVERWRITE_TAG == true ] ; then
    echo "Overwriting unpublished tag v$NEW_TAG_VERSION."
    git tag -s "v$NEW_TAG_VERSION" --force
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
    echo "Release v$NEW_TAG_VERSION"
    echo ""
    echo "Changes:"
    git log $LAST_TAG..HEAD --pretty="format:%h %s" --abbrev=10 || true
  ) > .git/TAG_EDITMSG

  bash -c "$editor .git/TAG_EDITMSG"
  git tag -s "v$NEW_TAG_VERSION" -F .git/TAG_EDITMSG
}

function publishTag() {
  git push -v
  tag
  git push origin "v$NEW_TAG_VERSION"
}

function pretag() {

  if [ ! $dryrun ]; then
    checkBranch
  fi

  checkWorkingDir
  #runTests
  echo "All tests and checks necessary for release passed."
  echo ""
}

function tagging() {
  echo ""
  echo Release announcements should use semantic versioning.

  if [ $LAST_TAG != "" ] ; then
    echo Last version seems to be $LAST_TAG
  fi

  read -p "New version: " NEW_TAG_VERSION < /dev/tty;
  # normalize by removing leading v (i.e., v0.0.1)
  NEW_TAG_VERSION=$(echo $NEW_TAG_VERSION | sed 's/^v//')

  # check if the version is tagged
  go run update/releasenotes/check/check.go $NEW_TAG_VERSION && ec=$? || ec=$?

  if [ ! $ec -eq 0 ] ; then
    >&2 echo "Add a release note to update/releasenotes/releasenotes.go and try again."
  fi

  return $ec
}

function release() {
#  runTestsOnDrone
  checkPublishedTag
  checkUnusedTag
  publishTag
}

function run() {
  CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
  LAST_TAG="$(git describe HEAD --tags --abbrev=0 2> /dev/null)" || true

  if [ $dryrun == true ]; then
    pretag
  else
    echo "Changes:"
    git --no-pager log $LAST_TAG..HEAD --pretty="format:%h %s" --abbrev=10 || true
    echo ""
    echo ""
    tagging
    pretag
    echo
    release
  fi
}

run
