#! /bin/sh

set -e

TAG=$1
DOCKERACC=$2

if [ -z "${TAG}" ]; then
	TAG="$(git rev-parse --short HEAD)"
	if [ -z "${TAG}" ]; then
		echo no TAG
		exit 2
	fi
	echo "No TAG specified. Using ${TAG}."
fi

if [ "${DOCKERACC}" = "" ]; then
	DOCKERACC=midonet
fi

VARIANTS="amd64-linux arm64v8-linux"

for VARIANT in $VARIANTS; do
        #
        # Partial build to run unit tests and fetch JUnit report file (despite succeded or not)
        #
        # Includes 'trick' (see https://stackoverflow.com/a/49891339) to
        # bypass Docker limitation about copy files from build-time containers.
        #
	docker build --target builder -t builder-${VARIANT} .
	BUILDER_CONTAINER="$(docker run -d builder-${VARIANT})"
	docker cp "$BUILDER_CONTAINER:/tmp/junit.xml" junit-${VARIANT}.xml
	echo "Generated JUnit XML report for '${VARIANT}' variant: junit-${VARIANT}.xml"
	docker rm -f "$BUILDER_CONTAINER"
done

for VARIANT in $VARIANTS; do
	docker build -f Dockerfile-${VARIANT} -t ${DOCKERACC}/midonet-kube-controllers-${VARIANT}:${TAG} .
	docker build -f Dockerfile-node-${VARIANT} -t ${DOCKERACC}/midonet-kube-node-${VARIANT}:${TAG} .
done

echo "Now you can push images with the following commands:"
for VARIANT in $VARIANTS; do
	echo "  docker push ${DOCKERACC}/midonet-kube-controllers-${VARIANT}:${TAG}"
	echo "  docker push ${DOCKERACC}/midonet-kube-node-${VARIANT}:${TAG}"
done
