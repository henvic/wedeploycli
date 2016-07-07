#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

VERSION_PKG='8WGbGy94JXa'

DESTDIR=/usr/local/bin
DEST=$DESTDIR/we
UNAME=$(uname)
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/i686/386/')
UNAME_ARCH=$(echo ${UNAME}_${ARCH} | tr '[:upper:]' '[:lower:]' | tr '_' '-')
FILE=cli-stable-$UNAME_ARCH
URL=https://bin.equinox.io/c/${VERSION_PKG}/$FILE.zip
echo Downloading $URL
curl -O $URL -f --progress-bar
unzip $FILE.zip -d $DESTDIR
rm $FILE.zip
chmod +x $DEST
