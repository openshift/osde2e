#!/bin/bash
#
# alert will look back at the previous job failures and generate granular alerts per-test-group
#

set -e

#docker pull quay.io/app-sre/osde2e
docker run -e PROMETHEUS_ADDRESS -e PROMETHEUS_BEARER_TOKEN -e SLACK_WEBHOOK quay.io/app-sre/osde2e alert