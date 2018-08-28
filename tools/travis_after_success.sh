#! /bin/sh

if [ "$DOCKER_BUILD" ]; then
	exit 0
else
	exec $GOPATH/bin/goveralls -coverprofile=profile.cov -service=travis-ci
fi
