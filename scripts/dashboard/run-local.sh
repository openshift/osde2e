#!/bin/bash
set -e

echo "Starting osde2e Dashboard locally..."
echo ""
echo "Make sure you have OCM credentials set:"
echo "  export OCM_CLIENT_ID=xxx"
echo "  export OCM_CLIENT_SECRET=yyy"
echo "  export OCM_ENV=stage (or prod)"
echo ""

# Build
echo "Building osde2e..."
go build -o bin/osde2e ./cmd/osde2e

# Run dashboard
echo "Starting dashboard on http://localhost:8080"
./bin/osde2e dashboard --environment stage --port 8080
