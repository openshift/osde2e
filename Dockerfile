FROM golang:1.11.9-alpine3.9
ENV PKG=/go/src/github.com/openshift/osde2e/
WORKDIR ${PKG}

# install build prerequisites
RUN apk add --no-cache make glide git

# resolve and install imports
ADD glide.yaml glide.lock ${PKG}
RUN glide install --strip-vendor

# compile test binary
ADD . ${PKG}
RUN make out/osde2e

# run tests
CMD ["make", "test"]
