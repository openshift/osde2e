#!/bin/bash
# Deploys the Delivery Dashboard to the delivery-dashboard namespace
# on the currently logged-in OpenShift cluster.
#
# Prerequisites:
#   - docker login quay.io (push credentials)
#   - oc login to target cluster
#   - Secrets pre-created in the namespace:
#       ocm-credentials  (keys: OCM_CLIENT_ID, OCM_CLIENT_SECRET)
#       aws-credentials  (keys: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
#   - hp-delivery-apps repo cloned adjacent to this repo
#     (git@gitlab.cee.redhat.com:hybrid-platforms-gitops/tenant-apps/hp-delivery-apps.git)
#
# Usage:
#   ./scripts/dashboard/deploy.sh [SQS_QUEUE_URL]
#
# Environment variables:
#   DASHBOARD_IMAGE   Image to build and deploy (default: quay.io/rmundhe_oc/delivery-dashboard:latest)
#   QUAY_EXPIRE       If set, adds quay.expires-after label (e.g. 26w). Use for dev/local builds.
#   SQS_QUEUE_URL     SQS queue URL for S3 event notifications
#   OVERLAY           Kustomize overlay to apply, relative to delivery-dashboard/ (default: overlays/stage)

set -euo pipefail

NAMESPACE="delivery-dashboard"
APP="delivery-dashboard"
IMAGE="${DASHBOARD_IMAGE:-quay.io/rmundhe_oc/delivery-dashboard:latest}"
QUAY_EXPIRE="${QUAY_EXPIRE:-}"
SQS_QUEUE_URL="${1:-${SQS_QUEUE_URL:-}}"
OVERLAY="${OVERLAY:-overlays/local}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MANIFESTS_REPO="$(cd "${REPO_ROOT}/../hp-delivery-apps" && pwd)"
OVERLAY_DIR="${MANIFESTS_REPO}/delivery-dashboard/${OVERLAY}"

echo "=== Delivery Dashboard Deployment ==="
echo "Namespace:  ${NAMESPACE}"
echo "Image:      ${IMAGE}"
echo "Overlay:    ${OVERLAY_DIR}"
echo "Cluster:    $(oc whoami --show-server)"
echo ""

# 1. Ensure namespace exists
oc new-project "${NAMESPACE}" 2>/dev/null || oc project "${NAMESPACE}"

# 2. Build container image locally and push to quay
echo "[1/4] Building and pushing container image..."
cd "${REPO_ROOT}"
EXPIRE_LABEL_ARG=""
if [[ -n "${QUAY_EXPIRE}" ]]; then
  EXPIRE_LABEL_ARG="--label quay.expires-after=${QUAY_EXPIRE}"
  echo "  (quay.expires-after=${QUAY_EXPIRE} will be applied)"
fi
# shellcheck disable=SC2086
docker build -f dashboard.Dockerfile ${EXPIRE_LABEL_ARG} -t "${IMAGE}" .
docker push "${IMAGE}"

# 3. Patch SQS_QUEUE_URL into the overlay configmap if provided, then apply via kustomize
echo "[2/4] Applying manifests via kustomize..."
if [[ -n "${SQS_QUEUE_URL}" ]]; then
  # Patch the configmap in a temp copy so we don't dirty the manifests repo
  TMPDIR=$(mktemp -d)
  trap 'rm -rf "${TMPDIR}"' EXIT
  cp -r "${OVERLAY_DIR}/." "${TMPDIR}/"
  # Update SQS_QUEUE_URL in the configmap
  sed -i.bak "s|SQS_QUEUE_URL:.*|SQS_QUEUE_URL: \"${SQS_QUEUE_URL}\"|" "${TMPDIR}/configmap.yaml"
  # Update image tag in kustomization
  (cd "${TMPDIR}" && kustomize edit set image "quay.io/rmundhe_oc/delivery-dashboard=${IMAGE}")
  kustomize build "${TMPDIR}" | oc apply -f -
else
  (cd "${OVERLAY_DIR}" && kustomize edit set image "quay.io/rmundhe_oc/delivery-dashboard=${IMAGE}")
  kustomize build "${OVERLAY_DIR}" | oc apply -f -
fi

echo "[3/4] Waiting for rollout..."
oc rollout status "deployment/${APP}" -n "${NAMESPACE}" --timeout=120s

echo "[4/4] Done!"
echo ""
ROUTE=$(oc get route "live" -n "${NAMESPACE}" -o jsonpath='{.spec.host}')
echo "Dashboard URL: https://${ROUTE}/dashboard/deliverables"