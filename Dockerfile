# Patching CVE-2026-32280, CVE-2026-32281, CVE-2026-32283, CVE-2026-33810: requires Go >= 1.26.2
# Patching CVE-2026-33811, CVE-2026-33814, CVE-2026-39817, CVE-2026-39819, CVE-2026-39820,
# CVE-2026-39823, CVE-2026-39825, CVE-2026-39826, CVE-2026-39836, CVE-2026-42499, CVE-2026-42501,
# CVE-2026-42504, CVE-2026-42507, CVE-2026-27145: requires Go >= 1.26.4
# Patching CVE-2026-39822, CVE-2026-42505: requires Go >= 1.26.5
# Patching CVE-2026-33818, CVE-2026-39821, CVE-2026-46600, CVE-2026-56853, CVE-2026-56858,
# CVE-2026-56859, CVE-2026-56860, CVE-2026-56862, CVE-2026-56864, CVE-2026-56865:
# requires Go >= 1.26.6
FROM golang:1.26.6 AS builder

RUN apt-get update && \
    dpkg --add-architecture arm64 &&\
    apt-get install -y --no-install-recommends build-essential && \
    apt-get clean && \
    mkdir -p "$GOPATH/src/github.com/bitnami-labs/kubewatch"

ADD . "$GOPATH/src/github.com/bitnami-labs/kubewatch"

RUN cd "$GOPATH/src/github.com/bitnami-labs/kubewatch" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=$(dpkg --print-architecture) go build -a --installsuffix cgo --ldflags="-s" -o /kubewatch

# Patching CVE-2026-4046, CVE-2026-4437: requires glibc >= 2.44, provided by chainguard/bash built after May 2026
# Patching CVE-2026-34180, CVE-2026-34181, CVE-2026-34182 (+12 more): requires libcrypto3 >= 3.6.3-r0
# Patching CVE-2026-58055: requires libnghttp2-14 >= 1.70.0-r0
# Both are pulled in by rebuilding on a current :latest base - this image must be rebuilt
# regularly (within 30 days) for the tag to stay clean, even when nothing else here changes.
FROM cgr.dev/chainguard/bash:latest

COPY --from=builder /kubewatch /bin/kubewatch

ENV KW_CONFIG=/opt/bitnami/kubewatch

ENTRYPOINT ["/bin/kubewatch"]
