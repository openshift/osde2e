#!/bin/bash
#
# Runs the SD scale test as defined in app-interface.
#

# Check the validity of the number of clusters
if ! echo "$NUMBER_OF_CLUSTERS" | grep -qE "^[0-9]+$"; then
  echo "Invalid number of clusters: number of clusters is malformed, should be a number."
  exit 1
fi

if [ "$NUMBER_OF_CLUSTERS" -le "0" ]; then
  echo "Invalid number of clusters: number of clusters needs to be greater than zero."
  exit 1
fi

# Burst tests don't need anything other than the number of clusters
if [ "$SCALE_TEST_TYPE" == "BURST" ]; then
  echo "Burst"

  # Set the size of batches to values that will provision everything at once.
  SIZE_OF_BATCHES=-1
  SECONDS_BETWEEN_BATCHES=1
# Concurrent management tests need to validate the size of batches and seconds between batches
elif [ "$SCALE_TEST_TYPE" == "CONCURRENT_MANAGEMENT" ]
then

  # We're just going to validate the values, then we'll break out of this to call the osde2ectl command.
  if ! echo "$SIZE_OF_BATCHES" | grep -qE "^[0-9]+$"; then
    echo "Invalid size of batches: size of batches is malformed, should be a number."
    exit 1
  fi

  if ! echo "$SECONDS_BETWEEN_BATCHES" | grep -qE "^[0-9]+$"; then
    echo "Invalid seconds between batches: seconds between batches is malformed, should be a number."
    exit 1
  fi

  if [ "$SIZE_OF_BATCHES" -le "0" ]; then
      echo "Invalid size of batches: size of batches needs to be greater than zero."
    exit 1
  fi

  if [ "$SECONDS_BETWEEN_BATCHES" -le "0" ]; then
    echo "Invalid seconds between batches: seconds between batches needs to be greater than zero."
    exit 1
  fi
else
  echo "Invalid scale test type."
  exit 1
fi

# Check the validity of the cluster expiry
if ! echo "$EXPIRY_IN_MINUTES" | grep -qE "^[0-9]+$"; then
  echo "Invalid expiry in minutes: expiry in minutes is malformed, should be a number."
  exit 1
fi

if [ "$EXPIRY_IN_MINUTES" -lt "0" ]; then
  echo "Invalid expiry in minutes: expiry in minutes needs to be greater than or equal to zero."
  exit 1
fi

echo "Running $SCALE_TEST_TYPE test..."

mkdir report

OSDE2ECTL=quay.io/app-sre/osde2ectl
docker pull $OSDE2ECTL

NUM_BATCHES=$(( (NUMBER_OF_CLUSTERS + (SIZE_OF_BATCHES - 1)) / SIZE_OF_BATCHES ))
ALL_CLUSTERS_PROVISIONED_SECONDS=$(( NUM_BATCHES * SECONDS_BETWEEN_BATCHES ))
TIME_UNTIL_ALL_CLUSTERS_EXPIRED=$(( ALL_CLUSTERS_PROVISIONED_SECONDS + EXPIRY_IN_MINUTES * 60 ))

if [ "$EXPIRY_IN_MINUTES" = "0" ]
then
  # If the expiry is set to 0, just use a hardcoded 6 hours.
  TIME_UNTIL_ALL_CLUSTERS_EXPIRED="21600"
fi

START_TIMESTAMP="$(date +%s%3N)"
END_TIMESTAMP="$(date -d "+$TIME_UNTIL_ALL_CLUSTERS_EXPIRED seconds" +%s%3N)"

echo "Check for results at https://grafana.app-sre.devshift.net/d/sd-scale/service-delivery-scale?from=$START_TIMESTAMP&to=$END_TIMESTAMP"

docker run -u "$(id -u)" -e OCM_TOKEN -e "REPORT_DIR=/report" -e "CLUSTER_EXPIRY_IN_MINUTES=$EXPIRY_IN_MINUTES" -e "OCM_USER_OVERRIDE=ci-ext-jenkins" -v "$(pwd)/report:/report" "$OSDE2ECTL" create --configs scale -n "$NUMBER_OF_CLUSTERS" -b "$SIZE_OF_BATCHES" -s "$SECONDS_BETWEEN_BATCHES"
