#! /bin/sh

set -e

TAG=$1
DOCKERACC=$2

if [ "${DOCKERACC}" = "" ]; then
	DOCKERACC=midonet
fi

docker build -f Dockerfile -t ${DOCKERACC}/midonet-kube-controllers .
docker build -f Dockerfile-node -t ${DOCKERACC}/midonet-kube-node .
docker tag ${DOCKERACC}/midonet-kube-controllers ${DOCKERACC}/midonet-kube-controllers:${TAG}
docker tag ${DOCKERACC}/midonet-kube-node ${DOCKERACC}/midonet-kube-node:${TAG}

echo "Now you can push images with the following commands:"
echo "  docker push ${DOCKERACC}/midonet-kube-controllers:${TAG}"
echo "  docker push ${DOCKERACC}/midonet-kube-node:${TAG}"

