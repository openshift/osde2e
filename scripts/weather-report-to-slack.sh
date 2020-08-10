#!/bin/bash
#
# gate-report will run the gate report command against the provided arguments and upload the resulting report to the osde2e-metrics bucket.
#

set -e

#docker pull quay.io/app-sre/osde2e
docker run -e REPORTING_WEATHER_JOB_ALLOWLIST -e PROMETHEUS_ADDRESS -e PROMETHEUS_BEARER_TOKEN -e SLACK_WEBHOOK quay.io/app-sre/osde2e report weather-report slack-summary --send-to-slack
