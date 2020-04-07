FROM registry.svc.ci.openshift.org/openshift/release:golang-1.13

ENV PKG=/go/src/github.com/openshift/osde2e/
WORKDIR ${PKG}

COPY . .
RUN make check
RUN make build

FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

COPY --from=0 /go/src/github.com/openshift/osde2e/out/osde2e .

ENTRYPOINT [ "/osde2e" ]
