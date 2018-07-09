ARG BUILD_WORKDIR=/go/src/github.com/midonet/midonet-kubernetes

FROM golang:1.10.3 as builder
ARG BUILD_WORKDIR
LABEL maintainer "YAMAMOTO Takashi <yamamoto@midokura.com>"
WORKDIR ${BUILD_WORKDIR}
COPY . .
RUN CGO_ENABLED=0 go build -ldflags '-s -w' -o midonet-kube-node ./cmd/midonet-kube-node
RUN CGO_ENABLED=0 go build -ldflags '-s -w' -o midonet-kube-cni ./cmd/midonet-kube-cni
RUN CGO_ENABLED=0 go build -ldflags '-s -w' -o midonet-kube-controllers ./cmd/midonet-kube-controllers

FROM scratch
LABEL maintainer "YAMAMOTO Takashi <yamamoto@midokura.com>"
ARG BUILD_WORKDIR
ARG BINARY
WORKDIR /root/
COPY --from=builder ${BUILD_WORKDIR}/midonet-kube-controllers .
CMD ["./midonet-kube-controllers"]
