#!/bin/bash -e

# Requires 3 successful passing builds before skipping
CLEAN_RUNS=3 make docker-test
