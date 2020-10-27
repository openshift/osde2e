#!/usr/bin/env bash

DIR=$(cd "$(dirname "$0")"/..; pwd)
OSDE2E=quay.io/app-sre/osde2e
HUGO=quay.io/jlelse/hugo:0.74.3

run_osde2e() {
	REPORTING_WEATHER_PROVIDER=$1 REPORTING_WEATHER_JOB_ALLOWLIST="osde2e-.*-$1-e2e-.*" docker run -u "$(id -u)" -v "$DIR:/hugo-site" -e REPORTING_WEATHER_PROVIDER -e REPORTING_WEATHER_JOB_ALLOWLIST -e PROMETHEUS_ADDRESS -e PROMETHEUS_BEARER_TOKEN "$OSDE2E" report weather-report sd-report --output "/hugo-site/content/post/$(uuidgen | sed s/-//g).md"
}

docker pull $OSDE2E

run_osde2e aws
run_osde2e gcp
run_osde2e moa

docker run -u $(id -u) -v "$DIR:/hugo-site" $HUGO hugo -s /hugo-site --cleanDestinationDir

git add "$DIR"
git commit -m "Weather report generation at $(date)"
