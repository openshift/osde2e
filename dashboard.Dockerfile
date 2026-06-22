FROM registry.access.redhat.com/ubi9/go-toolset:latest AS builder

USER root
ENV GOFLAGS=
ENV PKG=/opt/app-root/src/github.com/openshift/osde2e/
WORKDIR ${PKG}

COPY go.* .
RUN go mod download
COPY . .
RUN go env
RUN make build

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
WORKDIR /

COPY --from=builder /opt/app-root/src/github.com/openshift/osde2e/out/osde2e .

ENV PATH="${PATH}:/"
ENTRYPOINT ["/osde2e"]

LABEL name="delivery-dashboard"
LABEL description="Delivery Dashboard — pipeline status for Service Delivery operators, sourced from S3 and SQS"
LABEL summary="Web dashboard showing operator pipeline status across stage and integration environments"
LABEL com.redhat.component="delivery-dashboard"
LABEL io.k8s.description="delivery-dashboard"
LABEL io.k8s.display-name="Delivery Dashboard"
LABEL io.openshift.tags="dashboard,delivery,operators"
