#!/usr/bin/env bash

DIR=$(cd "$(dirname "$0")"/..; pwd)
#OSDE2E=quay.io/app-sre/osde2e
OSDE2E=8d2d6c0370c5

run_osde2e() {
	WEATHER_PROVIDER=$1 JOB_ALLOWLIST="osde2e-.*-$1-e2e-.*" docker run -u "$(id -u)" -v "$DIR:/hugo-site" -e WEATHER_PROVIDER -e JOB_ALLOWLIST -e PROMETHEUS_ADDRESS -e PROMETHEUS_BEARER_TOKEN "$OSDE2E" weather-report --output "/hugo-site/content/post/$(uuidgen | sed s/-//g).md" --outputType sd-report
}

run_osde2e aws
run_osde2e gcp
run_osde2e moa

docker run -u $(id -u) -v "$DIR:/hugo-site" klakegg/hugo:0.74.3-ext -s /hugo-site --cleanDestinationDir

git add "$DIR"
git commit -m "Weather report generation at $(date)"
git push
