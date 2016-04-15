#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

cd `dirname $0`
echo "Downloading linux kernel as test data"
curl http://cdn.kernel.org/pub/linux/kernel/v4.x/testing/linux-4.6-rc2.tar.xz -O

if [[ `openssl md5 linux-4.6-rc2.tar.xz` != *d13f04725958025e30e0f65a9f102e88* ]]; then
	echo "Failure to validate linux kernel package md5"
  exit 1
fi

echo "Decompressing"
tar xzf linux-4.6-rc2.tar.xz

echo "Removing compressed file"
rm linux-4.6-rc2.tar.xz
