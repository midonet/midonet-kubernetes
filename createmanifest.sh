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

for IMAGE in midonet-kube-controllers midonet-kube-node; do
	docker manifest create ${DOCKERACC}/${IMAGE}:${TAG} \
		${DOCKERACC}/${IMAGE}-amd64-linux:${TAG} \
		${DOCKERACC}/${IMAGE}-arm64v8-linux:${TAG}
	docker manifest annotate ${DOCKERACC}/${IMAGE}:${TAG} \
		${DOCKERACC}/${IMAGE}-amd64-linux:${TAG} \
		--arch amd64 --os linux
	docker manifest annotate ${DOCKERACC}/${IMAGE}:${TAG} \
		${DOCKERACC}/${IMAGE}-arm64v8-linux:${TAG} \
		--arch arm64 --os linux --variant v8
done

echo "Now you can push manifests with the following commands:"
echo "  docker manifest push ${DOCKERACC}/midonet-kube-controllers:${TAG}"
echo "  docker manifest push ${DOCKERACC}/midonet-kube-node:${TAG}"
