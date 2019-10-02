FROM golang:1.11.9
ENV PKG=/go/src/github.com/openshift/osde2e/
WORKDIR ${PKG}

# install build prerequisites
RUN apt-get update && apt-get install -y make golang-glide git

# resolve and install imports
ADD glide.yaml glide.lock ${PKG}
RUN glide install --strip-vendor

# compile test binary
ADD . ${PKG}

RUN make check

RUN make out/osde2e
RUN make out/osde2e-report

FROM gcr.io/distroless/base
COPY --from=0 /go/src/github.com/openshift/osde2e/out/osde2e .
COPY --from=0 /go/src/github.com/openshift/osde2e/out/osde2e-report .

ENTRYPOINT [ "/osde2e" ]