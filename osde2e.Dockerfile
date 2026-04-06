FROM registry.access.redhat.com/ubi9/ubi-minimal:latest AS podman-installer
RUN microdnf install -y podman-remote && microdnf clean all

FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_1.25 AS builder

ENV GOFLAGS=
ENV PKG=/go/src/github.com/openshift/osde2e/
WORKDIR ${PKG}

COPY go.* .
RUN go mod download
COPY . .
RUN go env
RUN make build

FROM registry.redhat.io/rhel9-2-els/rhel:9.2
WORKDIR /
# Create a writeable directory for licenses used by Tekton.
RUN mkdir /licenses

COPY --from=builder /go/src/github.com/openshift/osde2e/out/osde2e .
COPY --from=builder /go/src/github.com/openshift/osde2e/LICENSE /licenses/.
COPY --from=builder /usr/bin/git /usr/bin/git
COPY --from=builder /usr/libexec/git-core/* /usr/libexec/git-core/
COPY --from=builder /usr/share/git-core/* /usr/share/git-core/
COPY --from=podman-installer /usr/bin/podman-remote /usr/bin/podman-remote
COPY --from=podman-installer /usr/lib64/libsubid.so.3.0.0 /usr/lib64/libsubid.so.3.0.0
RUN ln -s /usr/bin/podman-remote /usr/bin/podman && \
    ln -s /usr/lib64/libsubid.so.3.0.0 /usr/lib64/libsubid.so.3

# Install OpenShift client (oc)
RUN curl -fsSL -o openshift-client-linux.tar.gz https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux.tar.gz \
    && tar -xzf openshift-client-linux.tar.gz \
    && mv oc /usr/local/bin/oc \
    && chmod +x /usr/local/bin/oc \
    && rm -f openshift-client-linux.tar.gz

ENV PATH="${PATH}:/"
ENTRYPOINT ["/osde2e"]
USER 65532:65532

LABEL name="osde2e"
LABEL description="A comprehensive test framework used for Service Delivery to test all aspects of Managed OpenShift Clusters"
LABEL summary="CLI tool to provision and test Managed OpenShift Clusters"
LABEL com.redhat.component="osde2e"
LABEL io.k8s.description="osde2e"
LABEL io.k8s.display-name="osde2e"
LABEL io.openshift.tags="data,images"
