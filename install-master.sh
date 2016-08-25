#!/bin/bash

# installs from gobuilder.me
# after building it there
# should take argument for tag/branch

set -euo pipefail
IFS=$'\n\t'

TAG="master"

UNAME=$(uname)
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/i686/386/')
UNAME_ARCH=$(echo ${UNAME}_${ARCH} | tr '[:upper:]' '[:lower:]' | tr '_' '-')
FILE=cli_${TAG}_$UNAME_ARCH
URL=https://gobuilder.me/get/github.com/wedeploy/cli/$FILE.zip

# default DESTDIR
DESTDIR=/usr/local/bin

function setupAlternateDir() {
  if [ ! -t 1 ] ; then
    echo "Can't install in $DESTDIR (default)."
    echo "Run this script from terminal to install it somewhere else."
    exit 1
  fi

  echo "No permission to install in $DESTDIR"
  echo "Cancel to run again as root / sudoer or install it somewhere else"
  read -p "Install in [current dir]: " DESTDIR < /dev/tty;
  DESTDIR=${DESTDIR:-`pwd`}
}

echo "CAUTION: Use at your own risk."
echo "This build is NOT stable and NOT intended for public use."
echo "Always download official releases from wedeploy.com or software updates"

if [ -t 1 ] ; then
  read -t 3 -p "" < /dev/tty || true
fi

if [ ! -w $DESTDIR ] ; then setupAlternateDir ; fi

function run() {
  echo Downloading from $URL
  curl -L -O $URL -f --progress-bar
  unzip -p $FILE.zip cli/cli > $DESTDIR/we
  chmod +x $DESTDIR/we
  info
}

function info() {
  wepath=$(which we)
  if [[ $wepath != "$DESTDIR/we" ]]; then
    echo "Installed from $TAG, but not on your \$PATH"
    echo "Run with .$DESTDIR/we"
    return
  fi

  we 2>&1 >/dev/null
  echo "Installed from $TAG, type 'we help' to start."
}

function cleanup() {
  rm $FILE.zip 2>/dev/null
}

trap cleanup EXIT
run
