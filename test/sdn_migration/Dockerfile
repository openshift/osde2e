FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_1.22  AS builder

ENV PKG=/go/src/github.com/openshift/osde2e/test/sdn_migration/
WORKDIR ${PKG}

# Copy go.mod and go.sum to the working directory
COPY ../../go.mod ../../go.sum ./

COPY ./test/sdn_migration/ .
RUN go mod tidy
RUN make build

FROM registry.redhat.io/rhel9-2-els/rhel:9.2
WORKDIR /

COPY --from=builder /go/src/github.com/openshift/osde2e/test/sdn_migration/out/sdn_migration .
ENTRYPOINT [ "/sdn_migration" ]
USER 65532:65532

LABEL name="sdn_migration"
LABEL description="A comprehensive test framework used for Service Delivery to validate rosa cluster network migration from SDN to OVN"
LABEL summary="CLI tool for validating the network migration of ROSA clusters"
LABEL com.redhat.component="sdn_migration"
LABEL io.k8s.description="sdn_migration"
LABEL io.k8s.display-name="sdn_migration"
LABEL io.openshift.tags="data,images"