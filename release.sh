#!/bin/bash

VERSION="0.1.0.1-dev"
NAME=micromdm

echo "Building $NAME version $VERSION"

mkdir -p build

build() {
  echo -n "=> $1-$2: "
  GOOS=$1 GOARCH=$2 CGO_ENABLED=0 go build -o build/$NAME-$1-$2 -ldflags "-X main.Version=$VERSION -X main.gitHash=`git rev-parse HEAD`" ./main.go
  du -h build/$NAME-$1-$2
}

build "darwin" "amd64"
build "linux" "amd64"
