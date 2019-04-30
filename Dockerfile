FROM golang:1.11.9-alpine3.9

RUN apk add --no-cache make glide git

WORKDIR /go/src/github.com/openshift/osde2e
ADD Makefile glide.yaml glide.lock /go/src/github.com/openshift/osde2e/
ADD ./cmd /go/src/github.com/openshift/osde2e/cmd
ADD ./pkg /go/src/github.com/openshift/osde2e/pkg

RUN glide install --strip-vendor
RUN make out/osde2e
ENTRYPOINT ["./out/osde2e"]
