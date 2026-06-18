FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
COPY osde2e /osde2e
ENTRYPOINT ["/osde2e"]
