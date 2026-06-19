FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
WORKDIR /
COPY osde2e /osde2e
ENV PATH="${PATH}:/"
ENTRYPOINT ["/osde2e"]

LABEL name="delivery-dashboard"
LABEL description="Delivery Dashboard — pipeline status for Service Delivery operators, sourced from S3 and SQS"
LABEL summary="Web dashboard showing operator pipeline status across stage and integration environments"
LABEL com.redhat.component="delivery-dashboard"
LABEL io.k8s.description="delivery-dashboard"
LABEL io.k8s.display-name="Delivery Dashboard"
LABEL io.openshift.tags="dashboard,delivery,operators"