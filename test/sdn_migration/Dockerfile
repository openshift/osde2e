FROM registry.ci.openshift.org/openshift/release:golang-1.21 AS builder
WORKDIR /tmp/src

COPY ../../go.mod go.sum ./

COPY . .

RUN make build

FROM registry.redhat.io/rhel9-2-els/rhel:9.2

COPY --from=builder /tmp/src/test/sdn_migration/*.test sdn_migration.test
RUN chmod +x sdn_migration.test
CMD [ "./sdn_migration.test" ]
