#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

VERSION_PKG='8WGbGy94JXa'

UNAME=$(uname)
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/i686/386/')
UNAME_ARCH=$(echo ${UNAME}_${ARCH} | tr '[:upper:]' '[:lower:]' | tr '_' '-')
FILE=cli-stable-$UNAME_ARCH
PACKAGE_FORMAT=""

# default DESTDIR
DESTDIR=/usr/local/bin

function setupAlternateDir() {
  if [ ! -t 1 ] ; then
    echo "Can't install in $DESTDIR (default)."
    echo "Run this script from terminal to install it somewhere else."
    exit 1
  fi

  echo "No permission to install in $DESTDIR"
  echo "Cancel to run again as root / sudoer or install it somewhere else."
  read -p "Install in [current dir]: " DESTDIR < /dev/tty;
  DESTDIR=${DESTDIR:-`pwd`}
}

if [ ! -w $DESTDIR ] ; then setupAlternateDir ; fi

function setPackageFormat() {
  (which tar >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    PACKAGE_FORMAT="tgz"
    return
  fi

  (which unzip >> /dev/null) && ec=$? || ec=$?

  if [ $ec -eq 0 ] ; then
    PACKAGE_FORMAT="zip"
    return
  fi

  >&2 echo "No available package format to install on your system."
  exit 1
}

function extractPackage() {
  case $PACKAGE_FORMAT in
    "zip")
    unzip -o $FILE.zip -d $DESTDIR we >/dev/null
    ;;
    "tgz")
    tar -xzf $FILE.$PACKAGE_FORMAT -C $DESTDIR we
    ;;
    *)
    echo "Error trying to extract binary from package."
    exit 1
    ;;
esac
}

function run() {
  setPackageFormat
  URL=https://bin.equinox.io/c/${VERSION_PKG}/$FILE.$PACKAGE_FORMAT
  echo Downloading from $URL
  curl -L -O $URL -f --progress-bar
  extractPackage
  chmod +x $DESTDIR/we
  info
}

function info() {
  wepath=(which we)
  if [[ $wepath != "$DESTDIR/we" ]]; then
    echo "Installed, but not on your \$PATH"
    echo "Run with $DESTDIR/we"
    return
  fi

  we 2>&1 >/dev/null
  echo "Installed, type 'we help' to start."
}

function cleanup() {
  rm $FILE.$PACKAGE_FORMAT 2>/dev/null
}

trap cleanup EXIT
run
