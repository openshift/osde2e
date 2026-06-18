FROM docker.io/golang:1.25 AS builder

ENV GOFLAGS="-mod=mod"
ENV PKG=/go/src/github.com/openshift/osde2e/
WORKDIR ${PKG}

COPY . .
RUN make build

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
WORKDIR /
COPY --from=builder /go/src/github.com/openshift/osde2e/out/osde2e .

ENV PATH="${PATH}:/"
ENTRYPOINT ["/osde2e"]

LABEL name="osde2e"
LABEL description="A comprehensive test framework used for Service Delivery to test all aspects of Managed OpenShift Clusters"
LABEL summary="CLI tool to provision and test Managed OpenShift Clusters"
