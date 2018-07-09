#! /bin/sh

set -e

TAG=$1
DOCKERACC=$2
IMAGES="midonet-kube-controllers midonet-kube-node"

for IMAGE in $IMAGES; do
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
for IMAGE in $IMAGES; do
	echo "  docker manifest push ${DOCKERACC}/${IMAGE}:${TAG}"
done
