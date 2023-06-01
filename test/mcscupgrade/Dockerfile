FROM registry.ci.openshift.org/openshift/release:golang-1.19 AS builder
WORKDIR /tmp/src
COPY . .
RUN make build

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
COPY --from=builder /tmp/src/*.test mcscupgrade.test
ENTRYPOINT [ "/mcscupgrade.test" ]