FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_1.22

ENV GOFLAGS=
ENV PKG=/go/src/github.com/openshift/osde2e/
WORKDIR ${PKG}

COPY . .
RUN go env
RUN make build

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

RUN microdnf install -y git && microdnf clean all
RUN mkdir /osde2e-bin
COPY --from=0 /go/src/github.com/openshift/osde2e/out/osde2e /osde2e-bin

# Restore the /osde2e path for backwards compatibility
RUN ln -s /osde2e-bin/osde2e /osde2e
ENV PATH "/osde2e-bin:$PATH"

ENTRYPOINT [ "osde2e" ]

LABEL name="osde2e"
LABEL description="A comprehensive test framework used for Service Delivery to test all aspects of Managed OpenShift Clusters"
LABEL summary="CLI tool to provision and test Managed OpenShift Clusters"
LABEL com.redhat.component="osde2e"
LABEL io.k8s.description="osde2e"
LABEL io.k8s.display-name="osde2e"
LABEL io.openshift.tags="data,images"
