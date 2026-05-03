# Patching CVE-2026-32280, CVE-2026-32281, CVE-2026-32283, CVE-2026-33810: requires Go >= 1.26.2
FROM golang:1.26.2 AS builder

RUN apt-get update && \
    dpkg --add-architecture arm64 &&\
    apt-get install -y --no-install-recommends build-essential && \
    apt-get clean && \
    mkdir -p "$GOPATH/src/github.com/bitnami-labs/kubewatch"

ADD . "$GOPATH/src/github.com/bitnami-labs/kubewatch"

RUN cd "$GOPATH/src/github.com/bitnami-labs/kubewatch" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=$(dpkg --print-architecture) go build -a --installsuffix cgo --ldflags="-s" -o /kubewatch

# Patching CVE-2026-4046, CVE-2026-4437: requires glibc >= 2.44, provided by chainguard/bash built after May 2026
FROM cgr.dev/chainguard/bash:latest

COPY --from=builder /kubewatch /bin/kubewatch

ENV KW_CONFIG=/opt/bitnami/kubewatch

ENTRYPOINT ["/bin/kubewatch"]
