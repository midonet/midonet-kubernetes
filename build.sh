#! /bin/sh

set -e

TAG=$1
DOCKERACC=$2

./generate-dockerfiles.sh check

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

docker build -f Dockerfile-test --target builder -t tester .
TESTER_CONTAINER="$(docker run -d tester)"
docker cp "$TESTER_CONTAINER:/tmp/junit.xml" .
echo "Generated JUnit XML report: junit.xml"
docker rm -f "$TESTER_CONTAINER"

VARIANTS="amd64-linux arm64v8-linux"

for VARIANT in $VARIANTS; do
	docker build -f Dockerfile-${VARIANT} -t ${DOCKERACC}/midonet-kube-controllers-${VARIANT}:${TAG} .
	docker build -f Dockerfile-node-${VARIANT} -t ${DOCKERACC}/midonet-kube-node-${VARIANT}:${TAG} .
done

echo "Now you can push images with the following commands:"
for VARIANT in $VARIANTS; do
	echo "  docker push ${DOCKERACC}/midonet-kube-controllers-${VARIANT}:${TAG}"
	echo "  docker push ${DOCKERACC}/midonet-kube-node-${VARIANT}:${TAG}"
done
