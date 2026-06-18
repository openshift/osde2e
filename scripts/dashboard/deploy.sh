#!/bin/bash
# Deploys the Delivery Dashboard to the delivery-dashboard namespace
# on the currently logged-in OpenShift cluster.
#
# Prerequisites:
#   - oc login to target cluster
#   - Secrets already exist: ocm-token, aws-credentials
#   - SQS_QUEUE_URL set (or passed as first arg)
#
# Usage:
#   ./scripts/dashboard/deploy.sh [SQS_QUEUE_URL]

set -euo pipefail

NAMESPACE="delivery-dashboard"
APP="delivery-dashboard"
SQS_QUEUE_URL="${1:-${SQS_QUEUE_URL:-}}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "=== Delivery Dashboard Deployment ==="
echo "Namespace: ${NAMESPACE}"
echo "Cluster:   $(oc whoami --show-server)"
echo ""

# 1. Ensure namespace exists
oc new-project "${NAMESPACE}" 2>/dev/null || oc project "${NAMESPACE}"

# 2. Build binary locally
echo "[1/5] Building osde2e binary..."
cd "${REPO_ROOT}"
GOFLAGS="-mod=mod" go build -o osde2e ./cmd/osde2e/

# 3. Build container image in cluster
echo "[2/5] Building container image..."
mkdir -p /tmp/dashboard-build
cp "${REPO_ROOT}/osde2e" /tmp/dashboard-build/osde2e
cp "${REPO_ROOT}/Dockerfile" /tmp/dashboard-build/Dockerfile

# Create BuildConfig if it doesn't exist
oc get buildconfig "${APP}" -n "${NAMESPACE}" &>/dev/null || \
  oc new-build --name="${APP}" --binary --strategy=docker -n "${NAMESPACE}"

oc start-build "${APP}" \
  --from-dir=/tmp/dashboard-build \
  --follow \
  -n "${NAMESPACE}"

# 4. Apply manifests
echo "[3/5] Applying manifests..."

# PVC for SQLite database
oc apply -n "${NAMESPACE}" -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: dashboard-db
  namespace: ${NAMESPACE}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF

# ConfigMap for non-secret config
oc apply -n "${NAMESPACE}" -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: dashboard-config
  namespace: ${NAMESPACE}
data:
  SQS_QUEUE_URL: "${SQS_QUEUE_URL}"
  LOG_BUCKET: "osde2e-logs"
  AWS_REGION: "us-east-1"
EOF

# Deployment
oc apply -n "${NAMESPACE}" -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${APP}
  namespace: ${NAMESPACE}
  labels:
    app: ${APP}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${APP}
  template:
    metadata:
      labels:
        app: ${APP}
    spec:
      containers:
        - name: dashboard
          image: image-registry.openshift-image-registry.svc:5000/${NAMESPACE}/${APP}:latest
          command: ["/osde2e"]
          args:
            - dashboard
            - --db=/data/dashboard.db
            - --backfill
            - --sqs-queue-url=\$(SQS_QUEUE_URL)
            - --port=8080
          env:
            - name: SQS_QUEUE_URL
              valueFrom:
                configMapKeyRef:
                  name: dashboard-config
                  key: SQS_QUEUE_URL
            - name: LOG_BUCKET
              valueFrom:
                configMapKeyRef:
                  name: dashboard-config
                  key: LOG_BUCKET
            - name: AWS_DEFAULT_REGION
              valueFrom:
                configMapKeyRef:
                  name: dashboard-config
                  key: AWS_REGION
          envFrom:
            - secretRef:
                name: ocm-token
            - secretRef:
                name: aws-credentials
          ports:
            - containerPort: 8080
          volumeMounts:
            - name: db
              mountPath: /data
          resources:
            requests:
              cpu: 100m
              memory: 256Mi
            limits:
              cpu: 500m
              memory: 512Mi
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 15
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 30
      volumes:
        - name: db
          persistentVolumeClaim:
            claimName: dashboard-db
EOF

# Service
oc apply -n "${NAMESPACE}" -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: ${APP}
  namespace: ${NAMESPACE}
spec:
  selector:
    app: ${APP}
  ports:
    - port: 8080
      targetPort: 8080
EOF

# Route
oc apply -n "${NAMESPACE}" -f - <<EOF
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: ${APP}
  namespace: ${NAMESPACE}
spec:
  to:
    kind: Service
    name: ${APP}
  port:
    targetPort: 8080
  tls:
    termination: edge
    insecureEdgeTerminationPolicy: Redirect
EOF

echo "[4/5] Waiting for rollout..."
oc rollout status deployment/${APP} -n "${NAMESPACE}" --timeout=120s

echo "[5/5] Done!"
echo ""
ROUTE=$(oc get route "${APP}" -n "${NAMESPACE}" -o jsonpath='{.spec.host}')
echo "Dashboard URL: https://${ROUTE}/dashboard/operators"
