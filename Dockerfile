###################################################
# Cloud-Barista CB-Dragonfly Module Dockerfile    #
###################################################

# Go 빌드 이미지 버전 및 알파인 OS 버전 정보
ARG BASE_IMAGE_BUILDER=golang
ARG GO_VERSION=1.14
ARG ALPINE_VERSION=3

###################################################
# 1. Build CB-Dragonfly binary file
###################################################

FROM ${BASE_IMAGE_BUILDER}:${GO_VERSION}-alpine AS go-builder

ENV CGO_ENABLED=0 \
	GO111MODULE="on" \
	GOOS="linux" \
	GOARCH="amd64" \
	GOPATH="/go/src/github.com/cloud-barista"

ARG GO_FLAGS="-mod=vendor"
ARG LD_FLAGS="-s -w"
ARG OUTPUT="bin/cb-dragonfly"

WORKDIR ${GOPATH}/cb-dragonfly
COPY . ./
RUN go build ${GO_FLAGS} -ldflags "${LD_FLAGS}" -o ${OUTPUT} -i ./pkg/manager/main \
    && chmod +x ${OUTPUT}

###################################################
# 2. Set up CB-Dragonfly runtime environment
###################################################

FROM alpine:${ALPINE_VERSION} AS runtime-alpine

ENV TZ="Asia/Seoul"

RUN apk add --no-cache \
    bash \
    tzdata \
    && \
    cp --remove-destination /usr/share/zoneinfo/${TZ} /etc/localtime \
    && \
    echo "${TZ}" > /etc/timezone

###################################################
# 3. Execute CB-Dragonfly Module
###################################################

FROM runtime-alpine as cb-dragonfly
LABEL maintainer="innogrid <dev.cloudbarista@innogrid.com>"

ENV GOPATH="/go/src/github.com/cloud-barista" \
    CBSTORE_ROOT=${GOPATH}/cb-dragonfly \
    CBLOG_ROOT=${GOPATH}/cb-dragonfly \
    CBMON_ROOT=${GOPATH}/cb-dragonfly

COPY --from=go-builder ${GOPATH}/cb-dragonfly/file ${GOPATH}/cb-dragonfly/file

WORKDIR /opt/cb-dragonfly
COPY --from=go-builder ${GOPATH}/cb-dragonfly/bin/cb-dragonfly /opt/cb-dragonfly/bin/cb-dragonfly
RUN chmod +x /opt/cb-dragonfly/bin/cb-dragonfly \
    && ln -s /opt/cb-dragonfly/bin/cb-dragonfly /usr/bin

EXPOSE 8094/udp
EXPOSE 9090

ENTRYPOINT ["cb-dragonfly"]
