#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

RELEASE_CHANNEL=${1:-"stable"}
RELEASE_CHANNEL_ADDRESS=""

if [[ $RELEASE_CHANNEL == "help" ]] || [[ $RELEASE_CHANNEL == "--help" ]]; then
  echo "WeDeploy CLI install script:

install.sh [channel] [dest]

Use install.sh to install the stable version on your system."
  exit 1
fi

UNAME=$(uname)
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/i686/386/')
UNAME_ARCH=$(echo ${UNAME}_${ARCH} | tr '[:upper:]' '[:lower:]' | tr '_' '-')
FILE=cli-$RELEASE_CHANNEL-$UNAME_ARCH
PACKAGE_FORMAT=""
DESTDIR=${2:-"/usr/local/bin"}

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
  DESTDIR=${DESTDIR/"~"/$HOME}
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

function setupReleaseChannelAddress() {
  case $RELEASE_CHANNEL in
    "stable") RELEASE_CHANNEL_ADDRESS="8WGbGy94JXa" ;;
    "unstable") RELEASE_CHANNEL_ADDRESS="5VvYPvs2CSX" ;;
    *) echo "Error translating release channel glob." exit 1 ;;
  esac
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
  setupReleaseChannelAddress
  URL=https://bin.equinox.io/c/${RELEASE_CHANNEL_ADDRESS}/$FILE.$PACKAGE_FORMAT

  if [ $RELEASE_CHANNEL != "stable" ] ; then
    echo "Downloading from $URL ($RELEASE_CHANNEL channel)."
  else
    echo "Downloading from $URL."
  fi

  curl -L -O $URL -f --progress-bar
  extractPackage
  chmod +x $DESTDIR/we
  info
}

function info() {
  wepath=`which we`
  if [[ ! $wepath -ef "$DESTDIR/we" ]]; then
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
