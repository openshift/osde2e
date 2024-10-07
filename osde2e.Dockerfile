FROM brew.registry.redhat.io/rh-osbs/openshift-golang-builder:rhel_9_1.22 as builder

ENV GOFLAGS=
ENV PKG=/go/src/github.com/openshift/osde2e/
WORKDIR ${PKG}

COPY . .
RUN go env
RUN make build

FROM registry.redhat.io/rhel9-2-els/rhel:9.2
WORKDIR /
RUN mkdir /licenses
RUN mkdir /osde2e-bin
COPY --from=builder /go/src/github.com/openshift/osde2e/out/osde2e /osde2e-bin
COPY --from=builder /go/src/github.com/openshift/osde2e/LICENSE /licenses/.
COPY --from=builder /usr/bin/git /usr/bin/git
COPY --from=builder /usr/libexec/git-core/* /usr/libexec/git-core/
COPY --from=builder /usr/share/git-core/* /usr/share/git-core/

RUN ln -s /osde2e-bin/osde2e /osde2e
ENV PATH "/osde2e-bin:$PATH"

ENTRYPOINT ["/osde2e"]
USER 65532:65532

LABEL name="osde2e"
LABEL description="A comprehensive test framework used for Service Delivery to test all aspects of Managed OpenShift Clusters"
LABEL summary="CLI tool to provision and test Managed OpenShift Clusters"
LABEL com.redhat.component="osde2e"
LABEL io.k8s.description="osde2e"
LABEL io.k8s.display-name="osde2e"
LABEL io.openshift.tags="data,images"
