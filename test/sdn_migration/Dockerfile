FROM registry.ci.openshift.org/openshift/release:golang-1.22 AS builder

ENV PKG=/go/src/github.com/openshift/osde2e/test/sdn_migration/
WORKDIR ${PKG}

# Copy go.mod and go.sum to the working directory
COPY ../../go.mod ../../go.sum ./

COPY ./test/sdn_migration/ .
RUN make build

FROM registry.redhat.io/rhel9-2-els/rhel:9.2
WORKDIR /

COPY --from=builder /go/src/github.com/openshift/osde2e/test/sdn_migration/out/sdn_migration .
ENTRYPOINT [ "/sdn_migration" ]
