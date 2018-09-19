include(Dockerfile.arm64v8-linux)
FROM arm64v8/alpine:3.8
LABEL maintainer "YAMAMOTO Takashi <yamamoto@midokura.com>"
ARG BUILD_WORKDIR
WORKDIR /root/
COPY node-scripts .
COPY --from=builder ${BUILD_WORKDIR}/dist/arm64-linux/midonet-kube-node .
COPY --from=builder ${BUILD_WORKDIR}/dist/arm64-linux/midonet-kube-cni .
CMD ["./main"]
