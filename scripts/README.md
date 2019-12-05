# Scripts

Various scripts that support the execution and day-to-day operation of osde2e.

## metrics-sync.sh

This script syncs the prometheus export files located in the osde2e-metrics S3 bucket with the Datahub prometheus instance via Datahub prometheus pushgateway.

Right now this script depends on python3+ and virtualenv in order to run, simply to setup and install the latest awscli.
