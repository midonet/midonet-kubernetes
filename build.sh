#! /bin/sh

set -e

TAG=$1

docker build -f Dockerfile -t midonet/midonet-kube-controllers .
docker build -f Dockerfile-node -t midonet/midonet-kube-node .

echo "Now you can tag and push images with the following commands:"
echo "  docker tag midonet/midonet-kube-controllers midonet/midonet-kube-controllers:${TAG}"
echo "  docker tag midonet/midonet-kube-node midonet/midonet-kube-node:${TAG}"
echo "  docker push midonet/midonet-kube-controllers:${TAG}"
echo "  docker push midonet/midonet-kube-node:${TAG}"

