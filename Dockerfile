FROM golang:1.11.9
ENV PKG=/go/src/github.com/openshift/osde2e/
WORKDIR ${PKG}

# install build prerequisites
RUN apt-get update && apt-get install -y make golang-glide git

# resolve and install imports
COPY glide.yaml glide.lock ${PKG}
RUN glide install --strip-vendor

# compile test binary
COPY . ${PKG}

RUN make check

RUN make out/osde2e

FROM gcr.io/distroless/base

COPY artifacts artifacts
COPY --from=0 /go/src/github.com/openshift/osde2e/out/osde2e .

ENTRYPOINT [ "/osde2e" ]
