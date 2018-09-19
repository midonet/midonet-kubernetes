include(Dockerfile.amd64-linux)
FROM scratch
LABEL maintainer "YAMAMOTO Takashi <yamamoto@midokura.com>"
ARG BUILD_WORKDIR
ARG BINARY
WORKDIR /root/
COPY --from=builder ${BUILD_WORKDIR}/dist/amd64-linux/midonet-kube-controllers .
CMD ["./midonet-kube-controllers"]
