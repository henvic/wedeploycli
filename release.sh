#!/bin/bash

# Launchpad CLI tool publishing script

set -euo pipefail
IFS=$'\n\t'

# "/Users/henvic/projects/launchpad/equinox.yml"

config=${1-}

if [ -z $config ] || [ $config == "help" ] || [ $config == "--help" ]; then
	echo "Launchpad CLI Tool publishing script:

1) check if all changes are commited
2) run tests on a local drone.io instance
3) create and push a release tag
5) build and push a new release to equinox

Check Semantic Versioning rules on semver.org

Use ./release.sh <equinox config>"
	exit 1
fi

if [ `git status --short | wc -l` -gt 0 ]; then
	>&2 echo "Your changes are not commited."
	exit 1
fi

echo "Release announcements should start with v<version>.

You may want to verify the differences between two versions with
git log HEAD..(past version) --pretty=\"format:%s https://github.com/launchpad-project/cli/commit/%h\"

To create a summary of changes.
"

read -p "Release channel [stable]: " RELEASE_CHANNEL;
RELEASE_CHANNEL=${RELEASE_CHANNEL:-"stable"}

read -p "New version: " NEW_RELEASE_VERSION;
# normalize by removing leading v (i.e., v0.0.1)
NEW_RELEASE_VERSION=`echo $NEW_RELEASE_VERSION | sed 's/^v//'

check_tags_free=`git tag --list "v$NEW_RELEASE_VERSION" | wc -l`
if [ "$check_tags_free" -gt 0 ]; then
	>&2 echo "git tag v$NEW_RELEASE_VERSION already exists"
	exit 1
fi

git tag -s "v$NEW_RELEASE_VERSION"
git checkout "v$NEW_RELEASE_VERSION"

drone exec

git push origin "v$NEW_RELEASE_VERSION"

equinox release \
--version=$NEW_RELEASE_VERSION \
--channel=$RELEASE_CHANNEL \
--config=$config \
-- \
-ldflags="-X github.com/launchpad-project/cli/defaults.Version=$NEW_RELEASE_VERSION" \
github.com/launchpad-project/cli
