#! /bin/sh

set -e

if [ "$DOCKER_BUILD" ]; then
	exec ./build.sh TEST
else
	go mod init
	go get github.com/mattn/goveralls
	exec ./tools/check-all.sh
fi
