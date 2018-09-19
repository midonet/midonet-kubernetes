include(Dockerfile.amd64-linux)
FROM amd64/alpine:3.8
LABEL maintainer "YAMAMOTO Takashi <yamamoto@midokura.com>"
ARG BUILD_WORKDIR
WORKDIR /root/
COPY node-scripts .
COPY --from=builder ${BUILD_WORKDIR}/dist/amd64-linux/midonet-kube-node .
COPY --from=builder ${BUILD_WORKDIR}/dist/amd64-linux/midonet-kube-cni .
CMD ["./main"]
