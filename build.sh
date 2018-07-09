#! /bin/sh

set -e

TAG=$1
DOCKERACC=$2

if [ "${DOCKERACC}" = "" ]; then
	DOCKERACC=midonet
fi

for VARIANT in amd64-linux arm64v8-linux; do
	docker build -f Dockerfile-${VARIANT} -t ${DOCKERACC}/midonet-kube-controllers-${VARIANT} .
	docker build -f Dockerfile-node-${VARIANT} -t ${DOCKERACC}/midonet-kube-node-${VARIANT} .
	docker tag ${DOCKERACC}/midonet-kube-controllers-${VARIANT} ${DOCKERACC}/midonet-kube-controllers-${VARIANT}:${TAG}
	docker tag ${DOCKERACC}/midonet-kube-node-${VARIANT} ${DOCKERACC}/midonet-kube-node-${VARIANT}:${TAG}
done

echo "Now you can push images with the following commands:"
for VARIANT in amd64-linux arm64v8-linux; do
	echo "  docker push ${DOCKERACC}/midonet-kube-controllers-${VARIANT}:${TAG}"
	echo "  docker push ${DOCKERACC}/midonet-kube-node-${VARIANT}:${TAG}"
done
