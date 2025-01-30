FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_1.22 as builder

ENV GOFLAGS=
ENV PKG=/go/src/github.com/openshift/osde2e/test/fedramp
WORKDIR ${PKG}

# Copy go.mod and go.sum to the working directory
COPY ../../go.mod ../../go.sum ./
COPY ../../LICENSE .
COPY ./test/fedramp/ .
RUN go mod download

# Install the Ginkgo command-line tool
RUN go install github.com/onsi/ginkgo/v2/ginkgo@latest

# Build the test binary
RUN go test -c -o osde2e-fedramp.test

FROM registry.redhat.io/rhel9-2-els/rhel:9.2
WORKDIR /

# Create the necessary config directory for ocm
RUN mkdir -p /.config && chown -R 65532:65532 /.config
RUN mkdir /licenses
COPY --from=builder /go/bin/ginkgo /usr/local/bin/
COPY --from=builder /go/src/github.com/openshift/osde2e/test/fedramp/osde2e-fedramp.test .
COPY --from=builder /go/src/github.com/openshift/osde2e/test/fedramp/LICENSE /licenses/.

# Set the entry point to 'ginkgo'
ENTRYPOINT ["/usr/local/bin/ginkgo"]
USER 65532:65532

# Provide default arguments
CMD ["-v", "--label-filter=Fedramp", "--junit-report=/tmp/osde2e-report/junit.xml", "./osde2e-fedramp.test"]

LABEL name="osde2e-fedramp-tests"
LABEL description="A test framework for testing ROSA clusters on the FedRAMP environment"
LABEL summary="Test framework for ROSA clusters on the FedRAMP environment"
LABEL com.redhat.component="osde2e-fedramp-tests"
LABEL io.k8s.description="osde2e-fedramp-tests"
LABEL io.k8s.display-name="osde2e-fedramp-tests"
LABEL io.openshift.tags="data,images,tests"
