#!/bin/bash

WORKDIR=`dirname $0`
cd $WORKDIR

PROGRAM_NAME="hb"

mkdir -p output/
rm -rf output/*

command_exists() {
    command -v "$@" > /dev/null 2>&1
}

build () {
    OS=$1
    ARCH=$2
    TAG=$3
    SUFF=""
    if [[ $OS == "windows" ]]; then
        SUFF=".exe"
    fi
    echo "Build ${PROGRAM_NAME}_${OS}_${ARCH}${SUFF} ..."
    CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} go build -tags "${TAG}" --ldflags="${ldflags}" -o output/${PROGRAM_NAME}_${OS}_${ARCH}${SUFF}
    cd output/
    upx -9 ${PROGRAM_NAME}_${OS}_${ARCH}${SUFF}
    cd ..
}

if command_exists upx; then
    build linux amd64 "all"
    build linux 386 "all"
    build windows amd64 "all"
    build windows 386 "all"
    build darwin amd64 "all"
else
    echo "upx not found in PATH"
fi