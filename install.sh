#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

VERSION_PKG='8WGbGy94JXa'

UNAME=$(uname)
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/i686/386/')
UNAME_ARCH=$(echo ${UNAME}_${ARCH} | tr '[:upper:]' '[:lower:]' | tr '_' '-')
FILE=cli-stable-$UNAME_ARCH
URL=https://bin.equinox.io/c/${VERSION_PKG}/$FILE.zip

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

if [ ! -w $DESTDIR ] ; then setupAlternateDir ; fi

function run() {
  echo Downloading from $URL
  curl -L -O $URL -f --progress-bar
  unzip -o $FILE.zip -d $DESTDIR we >/dev/null
  chmod +x $DESTDIR/we
  info
}

function info() {
  wepath=(which we)
  if [[ $wepath != "$DESTDIR/we" ]]; then
    echo "Installed, but not on your \$PATH"
    echo "Run with .$DESTDIR/we"
    return
  fi

  echo "Installed, type 'we help' to start."
}

function cleanup() {
  rm $FILE.zip 2>/dev/null
}

trap cleanup EXIT
run
