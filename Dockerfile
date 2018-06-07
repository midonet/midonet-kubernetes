ARG BUILD_WORKDIR=/go/src/github.com/midonet/midonet-kubernetes
ARG BINARY=midonet-kube-controllers
ARG PACKAGE=./cmd/${BINARY}

FROM golang:1.10.1 as deps
ARG BUILD_WORKDIR
WORKDIR ${BUILD_WORKDIR}
COPY Gopkg.toml .
COPY Gopkg.lock .
RUN go get -v -u github.com/golang/dep/cmd/dep
RUN dep ensure -v -vendor-only

FROM golang:1.10.1 as builder
ARG BUILD_WORKDIR
ARG BINARY
ARG PACKAGE
LABEL maintainer "YAMAMOTO Takashi <yamamoto@midokura.com>"
WORKDIR ${BUILD_WORKDIR}
COPY pkg pkg
COPY cmd cmd
COPY --from=deps ${BUILD_WORKDIR}/vendor vendor
RUN CGO_ENABLED=0 go build -v -o ${BINARY} ${PACKAGE}

FROM scratch
LABEL maintainer "YAMAMOTO Takashi <yamamoto@midokura.com>"
ARG BUILD_WORKDIR
ARG BINARY
WORKDIR /root/
COPY --from=builder ${BUILD_WORKDIR}/${BINARY} main
CMD ["./main"]
