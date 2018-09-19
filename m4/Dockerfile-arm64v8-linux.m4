include(Dockerfile.arm64v8-linux)
FROM scratch
LABEL maintainer "YAMAMOTO Takashi <yamamoto@midokura.com>"
ARG BUILD_WORKDIR
ARG BINARY
WORKDIR /root/
COPY --from=builder ${BUILD_WORKDIR}/dist/arm64-linux/midonet-kube-controllers .
CMD ["./midonet-kube-controllers"]
