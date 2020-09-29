#!/usr/bin/env bash

set -ex

### You may need to populate these or define them in your own environment variables

# default_vaultAddress=
# default_vaultProxyHost=
# default_vaultIAPServiceAccount=
# default_vaultIAPClientID=
# default_gcsBucket=
# default_gcsFilename=
# default_binaryGCSBucket=
# default_binaryGCSPath=

### Don't change these, probably

buildDate=$(date -u "+%Y-%m-%d %H:%M:%S UTC")
buildHash=$(git rev-parse HEAD)
buildVersion=$(git tag -l --points-at HEAD 'v*')
buildOS=$(uname -s | tr 'A-Z' 'a-z')
buildInstallMethod=binary

cd src

if [ ! -z $1 ] && [ ! -z $2 ] ; then
	os=$1
	arch=$2
fi

if [ "$os" == "" ] || [ "$arch" == "" ] ; then
    buildCommand="env CGO_ENABLED=0 go build"
    buildOutput="../build/cog"
else
    buildCommand="env CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -a"
    buildOutput="../build/cog_${buildVersion}_${os}_${arch}/cog"
fi

${buildCommand}\
	-ldflags "-X 'main.buildInstallMethod=${buildInstallMethod}' \
	-X 'main.buildDate=${buildDate}' \
	-X 'main.buildHash=${buildHash}' \
	-X 'main.buildVersion=${buildVersion}' \
	-X 'main.buildOS=${buildOS}' \
    -X 'main.default_vaultAddress=${default_vaultAddress}' \
    -X 'main.default_vaultProxyHost=${default_vaultProxyHost}' \
    -X 'main.default_vaultIAPServiceAccount=${default_vaultIAPServiceAccount}' \
    -X 'main.default_vaultIAPClientID=${default_vaultIAPClientID}' \
    -X 'main.default_gcsBucket=${default_gcsBucket}' \
    -X 'main.default_gcsFilename=${default_gcsFilename}' \
    -X 'main.default_binaryGCSBucket=${default_binaryGCSBucket}' \
    -X 'main.default_binaryGCSPath=${default_binaryGCSPath}'" \
	-o ${buildOutput}

	if [ "$os" != "" ] && [ "$arch" != "" ] ; then
		cd ../build
		if [ "$os" == "darwin" ] ; then
			zip -r cog_${buildVersion}_${os}_${arch}.zip cog_${buildVersion}_${os}_${arch}/
		else
			tar zcf cog_${buildVersion}_${os}_${arch}.tgz cog_${buildVersion}_${os}_${arch}/
		fi
		cat <<EOF > cog_${buildVersion}_${os}_${arch}/cog.version
Version: $buildVersion
Build Date: $buildDate
Git Hash: $buildHash
Build OS: $buildOS
Install Method: $buildInstallMethod
EOF
		cd -
	fi
cd ..
