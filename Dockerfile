FROM golang:1.11.9-alpine3.9

RUN apk add --no-cache make glide git

ENV PKG=/go/src/github.com/openshift/osde2e/

WORKDIR ${PKG}
ADD glide.yaml glide.lock ${PKG}
RUN glide install --strip-vendor
ADD . /go/src/github.com/openshift/osde2e/

RUN make out/osde2e
CMD ["make", "test"]
