FROM registry.svc.ci.openshift.org/openshift/release:golang-1.13

ENV PKG=/go/src/github.com/openshift/osde2e/
ENV GOMODULE111=on
WORKDIR ${PKG}

# resolve and install imports
COPY go.mod ./
COPY go.sum ./

RUN go mod download

# compile test binary
COPY . .
RUN make check
RUN make build

FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

COPY artifacts artifacts
COPY --from=0 /go/src/github.com/openshift/osde2e/out/osde2e .
COPY --from=0 /go/src/github.com/openshift/osde2e/out/osde2e-scale .

ENTRYPOINT [ "/osde2e" ]
