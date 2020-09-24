#!/usr/bin/env bash

set -ex

# Make sure git is installed
if [ -z "$(which git)" ] ; then
	echo "Git is not installed. Quitting."
	exit 1
fi

# Fetch remote tags
git fetch --prune-tags --prune --tags

# Get the current branch
branch=$(git rev-parse --abbrev-ref HEAD)
if [ "$branch" != "master" ] ; then
	echo "You must release the master branch only. Current branch is: $branch"
	exit 1
fi

# Make sure we only have untracked files in the branch
if [ ! -z "$(git status --untracked-files=no --porcelain)" ]; then
	echo "There are uncommitted changes in tracked files. Quitting."
	exit 1
fi

# Make sure a version isn't already set
curr_version=$(git tag -l --points-at HEAD 'v*')
if [ "$curr_version" != "" ] ; then
        echo "A version is already set: $curr_version"
        exit 1
fi

if [ "$VERSION" != "" ] ; then
        # The user is attempting to specify their own tag
        # Accept v1.2.3 as a valid version
        if [[ $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] ; then
                tag=$VERSION
        # Or accept 1.2.3 as a valid version
        elif [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] ; then
                tag="v$VERSION"
        # Otherwise, complain and exit
        else
                echo "Tag must be formatted in semver: 1.2.3"
                exit 1
        fi
else
        # Figure out the tag via git
        remote=$(git remote -v | grep ^origin | head -n1 | awk '{print $2}')
        versions=$(git ls-remote --tags --sort=-creatordate $remote 'v*' | head -n1 | awk '{print $2}' | cut -d / -f 3)
        if [ "$versions" == "" ] ; then
                tag="v1.0.0"
        else
                front=$(echo "$versions" | cut -d . -f 1-2)
                back=$(echo "$versions" | cut -d . -f 3)
                let "next = $back + 1"
                tag="$front.$next"
        fi
fi

git tag -a -m $tag $tag
git push --tags
