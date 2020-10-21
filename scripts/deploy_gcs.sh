#!/usr/bin/env bash

set -ex

# Make sure git is installed
if [ -z "$(which git)" ] ; then
	echo "Git is not installed. Quitting."
	exit 1
fi

# Make sure gsutil is installed
if [ -z "$(which gsutil)" ] ; then
	echo "gsutil is not installed. Quitting."
	exit 1
fi

# Make sure GCS_BUCKET and GCS_PATH are set
if [ -z $GCS_BUCKET ] ; then
	echo "GCS_BUCKET is not set. Quitting."
	exit 1
fi

if [ -z $GCS_PATH ] ; then
	echo "GCS_PATH is not set. Quitting."
	exit 1
fi
echo go
exit

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

# Get the current version
curr_version=$(git tag -l --points-at HEAD 'v*')
if [ "$curr_version" == "" ] ; then
        echo "No version set"
        exit 1
fi

pwd=`pwd`
cd ..
mv ${pwd} cog-${curr_version}
tar -zcvf cog-${curr_version}.tar.gz cog-${curr_version}/{CHANGELOG.md,LICENSE,Makefile,README.md,docs,examples,images,scripts,src,terraform}
mv cog-${curr_version}.tar.gz cog-${curr_version}/build/cog-${curr_version}.tar.gz
mv cog-${curr_version} ${pwd}
cd ${pwd}

gsutil cp build/cog_${curr_version}_darwin_amd64/cog build/cog_${curr_version}_darwin_amd64/cog.version gs://${GCS_BUCKET}/${GCS_PATH}/darwin/
gsutil cp build/cog_${curr_version}_linux_amd64/cog build/cog_${curr_version}_linux_amd64/cog.version gs://${GCS_BUCKET}/${GCS_PATH}/linux/
gsutil cp build/cog-${curr_version}.tar.gz gs://${GCS_BUCKET}/${GCS_PATH}/source/