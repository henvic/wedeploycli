#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

RELEASE_CHANNEL=${1:-"stable"}
RELEASE_CHANNEL_ADDRESS=""

if [[ $RELEASE_CHANNEL == "help" ]] || [[ $RELEASE_CHANNEL == "--help" ]] || [[ $RELEASE_CHANNEL == "-h" ]]; then
  echo "WeDeploy CLI install script:

$0 [channel] [dest]

Use $0 to install the stable version on your system."
  exit 1
fi

UNAME=$(uname | tr '[:upper:]' '[:lower:]')

if [[ $UNAME == *"windows"* ]] || [[ $UNAME == *"mingw"* ]] || [[ $UNAME == *"cygwin"* ]] ; then
  UNAME="windows"
fi

ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/i686/386/')
UNAME_ARCH=$(echo ${UNAME}_${ARCH} | tr '[:upper:]' '[:lower:]' | tr '_' '-')
FILE=we-$RELEASE_CHANNEL-$UNAME_ARCH
PACKAGE_FORMAT=""

# Hacking mktemp's incompatible parameters on BSD and Linux
TEMPDEST=$(mktemp 2>/dev/null || mktemp -t 'wedeploy-cli')

ec=$(which we 2>/dev/null || true)

if [ ! -z $ec ] ; then
  DESTDIR=$(dirname $(which we))
elif [ $UNAME == "windows" ] ; then
  IS_MINGWIN=${MSYSTEM:-""}
  if [ $HOME != "" ] && [[ ! -z IS_MINGWIN ]] ; then
    DESTDIR="$HOME/AppData/Local/Programs/we/bin"
  elif [[ $HOMEDRIVE$HOMEPATH != "" ]] ; then
    DESTDIR="$HOMEDRIVE$HOMEPATH\AppData\Local\Programs\we\bin"
  else
    DESTDIR="$USERPROFILE\AppData\Local\Programs\we\bin"
  fi
elif [[ :$PATH: == *:"$HOME/.local/bin":* ]] ; then
  DESTDIR="$HOME/.local/bin"
elif [[ :$PATH: == *:"$HOME/bin":* ]] ; then
  DESTDIR="$HOME/bin"
else
  DESTDIR="/usr/local/bin"
fi

DESTDIR=${2:-$DESTDIR}

function setupAlternateDir() {
  if [ ! -t 1 ] ; then
    echo "Can't install in $DESTDIR (default).\n"
    echo "Your \$PATH locations:"
    echo "${PATH//:/\n}"
    echo "See https://en.wikipedia.org/wiki/PATH_(variable) for more info on \$PATH.\n"
    echo "Run this script from terminal to install it somewhere else."
    exit 1
  fi

  echo "No permission to install in $DESTDIR"
  echo "Try again as root or run:"
  echo "curl https://cdn.wedeploy.com/cli/latest/wedeploy.sh -sL | sudo bash"
  read -p "Install in [current dir]: " DESTDIR < /dev/tty;
  DESTDIR=${DESTDIR:-$(pwd)}
  DESTDIR=${DESTDIR/"~"/$HOME}
}

if [ ! -w $DESTDIR ] ; then
  mkdir -p $DESTDIR
fi

if [ ! -w $DESTDIR ] ; then setupAlternateDir ; fi

echo "Trying to install in $DESTDIR"

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
    unzip -o $TEMPDEST -d $DESTDIR we >/dev/null
    ;;
    "tgz")
    tar -xzf $TEMPDEST -C $DESTDIR we
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

  curl -L -o $TEMPDEST $URL -f --progress-bar
  extractPackage
  chmod +x $DESTDIR/we
  info
}

function info() {
  wepath=$(which we 2>/dev/null)
  if [[ ! $wepath -ef "$DESTDIR/we" ]]; then
    echo "Installed, but not reachable by \"we\" (check your \$PATH)"
    echo "Run with $DESTDIR/we"
    return
  fi

  UNPRIVILEGED_USER=${SUDO_USER:-""}
  fixChannel
  check

  echo "Installed, type 'we help' to start."
}

function check() {
  if [ -z "$UNPRIVILEGED_USER" ]; then
    we 2>&1 >/dev/null
  else
    sudo --user $UNPRIVILEGED_USER we 2>&1 >/dev/null
  fi
}

# Verify direct installation of non-stable version
# and fix channel on ~/.we configuration file automatically.
# Fix issue https://github.com/wedeploy/cli/issues/472
function fixChannel() {
  if [[ $RELEASE_CHANNEL == "stable" ]]; then
    return
  fi

  if [ -z "$UNPRIVILEGED_USER" ]; then
    we update --channel $RELEASE_CHANNEL 2>&1 >/dev/null
  else
    sudo --user $UNPRIVILEGED_USER we update --channel $RELEASE_CHANNEL 2>&1 >/dev/null
  fi
}

function cleanup() {
  rm $FILE.$PACKAGE_FORMAT 2>/dev/null || true
  rm $TEMPDEST 2>/dev/null || true
}

trap cleanup EXIT
run
