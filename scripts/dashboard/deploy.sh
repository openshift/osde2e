#!/bin/bash
# Local dev deploy — builds image from source, pushes to quay, applies kustomize overlay.
# Not used by CI/prod. Set SQS_QUEUE_URL in overlays/local/configmap.yaml in hp-delivery-apps.
#
# Usage: DASHBOARD_QUAY_IMAGE=quay.io/<user>/delivery-dashboard:latest ./scripts/dashboard/deploy.sh
# Env:   DASHBOARD_QUAY_IMAGE (required), QUAY_EXPIRE (e.g. 26w), OVERLAY (default: overlays/local)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

[[ -z "${DASHBOARD_QUAY_IMAGE:-}" ]] && { echo "Error: DASHBOARD_QUAY_IMAGE is not set."; exit 1; }

IMAGE="${DASHBOARD_QUAY_IMAGE}"
QUAY_EXPIRE="${QUAY_EXPIRE:-}"
OVERLAY="${OVERLAY:-overlays/local}"
OVERLAY_DIR="$(cd "${REPO_ROOT}/../hp-delivery-apps" && pwd)/delivery-dashboard/${OVERLAY}"
NAMESPACE=$(grep "^namespace:" "${OVERLAY_DIR}/kustomization.yaml" | awk '{print $2}')
APP="delivery-dashboard"
BUILD_CTX="${REPO_ROOT}/configs/local/dashboard-build"

echo "=== Delivery Dashboard Deployment ==="
echo "Overlay:    ${OVERLAY}  (namespace: ${NAMESPACE})"
echo "Image:      ${IMAGE}"
echo "Cluster:    $(oc whoami --show-server)"
echo ""

oc new-project "${NAMESPACE}" 2>/dev/null || oc project "${NAMESPACE}"

echo "Checking secrets..."
MISSING=0
for SECRET in osde2e-ocm-credentials osde2e-aws-credentials; do
  oc get secret "${SECRET}" -n "${NAMESPACE}" &>/dev/null \
    && echo "  OK: ${SECRET}" \
    || { echo "  MISSING: ${SECRET}"; MISSING=1; }
done
[[ "${MISSING}" -eq 1 ]] && { echo "Create missing secrets first (see hp-delivery-apps/delivery-dashboard/README.md)"; exit 1; }

echo "[1/4] Building image..."
GOOS=linux GOARCH=amd64 GOFLAGS="-mod=mod" go build -o "${BUILD_CTX}/osde2e" "${REPO_ROOT}/cmd/osde2e/"
EXPIRE_ARG="${QUAY_EXPIRE:+--label quay.expires-after=${QUAY_EXPIRE}}"
# shellcheck disable=SC2086
podman build ${EXPIRE_ARG} --platform linux/amd64 -t "${IMAGE}" "${BUILD_CTX}"

echo "[2/4] Pushing image..."
podman push "${IMAGE}"

echo "[3/4] Applying manifests..."
kustomize build "${OVERLAY_DIR}" | oc apply -f -
oc rollout restart "deployment/${APP}" -n "${NAMESPACE}"

echo "[4/4] Waiting for rollout..."
oc rollout status "deployment/${APP}" -n "${NAMESPACE}" --timeout=120s

echo ""
echo "Dashboard URL: https://$(oc get route live -n "${NAMESPACE}" -o jsonpath='{.spec.host}')/dashboard/pipelines"