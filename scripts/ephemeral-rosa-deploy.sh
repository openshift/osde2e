#!/usr/bin/env bash
# Shared bonfire ROSA HCP deploy for Prow jobs.
# Writes ${SHARED_DIR}/cluster-id from hub ROSAControlPlane.status.id.
#
# Usage:
#   source scripts/ephemeral-rosa-deploy.sh
#   ephemeral_rosa_deploy
#   # ... run osde2e against SHARED_DIR/cluster-id ...
#   ephemeral_rosa_release
#
# Or: ./scripts/ephemeral-rosa-deploy.sh deploy|release
set -euo pipefail

EPHEMERAL_CREDS_DIR="${EPHEMERAL_CREDS_DIR:-/usr/local/ci-secrets/ephemeral-cluster}"
SHARED_DIR="${SHARED_DIR:-${ARTIFACT_DIR}/shared}"
RESERVATION_DURATION="${RESERVATION_DURATION:-3h}"
DEPLOY_TIMEOUT="${DEPLOY_TIMEOUT:-1800}"
BONFIRE_VERSION="${BONFIRE_VERSION:->=4.18.0}"

EPHEMERAL_NS="${EPHEMERAL_NS:-}"

_ephemeral_rosa_install_bonfire() {
  echo "Installing bonfire (crc-bonfire${BONFIRE_VERSION})..."
  export LANG=en_US.UTF-8
  export LC_ALL=en_US.UTF-8
  python3 -m venv /tmp/bonfire-venv
  # shellcheck source=/dev/null
  source /tmp/bonfire-venv/bin/activate
  python3 -m pip install --quiet --upgrade pip setuptools wheel
  python3 -m pip install --quiet --upgrade "crc-bonfire${BONFIRE_VERSION}"
  export BONFIRE_BOT=true
  export BONFIRE_NS_REQUESTER="${JOB_NAME:-openshift-ci-osde2e}"
}

_ephemeral_rosa_hub_login() {
  test -f "${EPHEMERAL_CREDS_DIR}/oc-login-token"
  test -f "${EPHEMERAL_CREDS_DIR}/oc-login-server"
  OC_LOGIN_TOKEN="$(cat "${EPHEMERAL_CREDS_DIR}/oc-login-token")"
  OC_LOGIN_SERVER="$(cat "${EPHEMERAL_CREDS_DIR}/oc-login-server")"
  EPH_KUBECONFIG_DIR="/tmp/ephemeral-kube"
  EPH_KUBECONFIG="${EPH_KUBECONFIG_DIR}/config"
  rm -rf "${EPH_KUBECONFIG_DIR}"
  mkdir -p "${EPH_KUBECONFIG_DIR}"
  export KUBECONFIG="${EPH_KUBECONFIG}"
  set +x
  oc login --token="${OC_LOGIN_TOKEN}" --server="${OC_LOGIN_SERVER}" --insecure-skip-tls-verify=true >/dev/null
  set -x 2>/dev/null || true
}

_ephemeral_rosa_get_cluster_id_from_capi() {
  local namespace="$1"
  local control_plane="${namespace}-cluster-control-plane"
  oc wait "rosacontrolplane/${control_plane}" \
    -n "${namespace}" \
    --for=jsonpath='{.status.ready}'=true \
    --timeout="${DEPLOY_TIMEOUT}s"
  oc get "rosacontrolplane/${control_plane}" \
    -n "${namespace}" \
    -o jsonpath='{.status.id}'
}

ephemeral_rosa_deploy() {
  mkdir -p "${SHARED_DIR}"
  _ephemeral_rosa_install_bonfire
  _ephemeral_rosa_hub_login

  local deploy_log="${SHARED_DIR}/bonfire-deploy.log"
  bonfire deploy rosa \
    --duration "${RESERVATION_DURATION}" \
    --timeout "${DEPLOY_TIMEOUT}" 2>&1 | tee "${deploy_log}"

  EPHEMERAL_NS=$(grep -oE "namespace 'ephemeral-[^']+'" "${deploy_log}" | tail -1 | tr -d "'" | awk '{print $2}')
  if [[ -z "${EPHEMERAL_NS}" ]]; then
    EPHEMERAL_NS=$(bonfire namespace list --mine 2>/dev/null | awk '/ephemeral-/ {print $1; exit}')
  fi
  test -n "${EPHEMERAL_NS}"
  bonfire namespace describe "${EPHEMERAL_NS}" | grep -q 'ROSA Cluster configuration detected'

  local cluster_id
  cluster_id=$(_ephemeral_rosa_get_cluster_id_from_capi "${EPHEMERAL_NS}")
  test -n "${cluster_id}"
  echo -n "${cluster_id}" > "${SHARED_DIR}/cluster-id"
  echo "ephemeral namespace=${EPHEMERAL_NS} cluster_id=${cluster_id}"
  export EPHEMERAL_NS
  export CLUSTER_ID="${cluster_id}"
}

ephemeral_rosa_release() {
  if [[ -z "${EPHEMERAL_NS}" ]]; then
    return 0
  fi
  if command -v bonfire >/dev/null 2>&1; then
    bonfire namespace release "${EPHEMERAL_NS}" -f || true
  fi
  EPHEMERAL_NS=""
  export EPHEMERAL_NS
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  case "${1:-deploy}" in
    deploy) ephemeral_rosa_deploy ;;
    release) ephemeral_rosa_release ;;
    *)
      echo "usage: $0 deploy|release" >&2
      exit 1
      ;;
  esac
fi
