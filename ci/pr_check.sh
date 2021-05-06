#!/bin/bash -e

make build

CLUSTER_ID=1sAHFzPxMgXQoqjDErH7CTN2NBb \
OCM_TOKEN=$(cat /usr/local/osde2e-credentials/ocm-token) \
./out/osde2e --configs=prod,aws