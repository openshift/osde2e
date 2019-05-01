###############
# build image #
###############
FROM golang:1.11.9-alpine3.9
ENV PKG=/go/src/github.com/openshift/osde2e/
WORKDIR ${PKG}

# install build prerequisites
RUN apk add --no-cache make glide git

# resolve and install imports
ADD glide.yaml glide.lock ${PKG}
RUN glide install --strip-vendor

# compile test binary
ADD . /go/src/github.com/openshift/osde2e/
RUN make out/osde2e

##############
# test image #
##############
FROM alpine:3.9.3
ENV PKG=/go/src/github.com/openshift/osde2e/

# install test prequisites
RUN apk add --no-cache make git

# copy Makefile + test executable
COPY --from=0 ${PKG}/Makefile ${PKG}/out/osde2e /
ADD .git /
CMD ["make", "test"]
