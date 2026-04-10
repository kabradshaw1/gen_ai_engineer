#!/usr/bin/env bash
set -euo pipefail

# Deploy all services to Minikube
# Prerequisites: minikube running, kubectl configured
# Usage: ./k8s/deploy.sh
#
# Manifest discovery is directory-based: kubectl apply -f <dir>/
# applies every .yml in the directory. Adding a new manifest file
# to an existing directory is all that's needed — no script changes.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "==> Enabling NGINX Ingress Controller..."
minikube addons enable ingress 2>/dev/null || true

echo "==> Creating namespaces..."
kubectl apply -f "$SCRIPT_DIR/ai-services/namespace.yml"
kubectl apply -f "$REPO_DIR/java/k8s/namespace.yml"
kubectl apply -f "$SCRIPT_DIR/monitoring/namespace.yml"
kubectl apply -f "$REPO_DIR/go/k8s/namespace.yml"

echo "==> Applying secrets..."
if [ -f "$REPO_DIR/java/k8s/secrets/java-secrets.yml" ]; then
  kubectl apply -f "$REPO_DIR/java/k8s/secrets/java-secrets.yml"
else
  echo "    WARN: java-secrets.yml not found — copy java-secrets.yml.template and fill in values"
fi

if [ -f "$REPO_DIR/go/k8s/secrets/go-secrets.yml" ]; then
  kubectl apply -f "$REPO_DIR/go/k8s/secrets/go-secrets.yml"
else
  echo "    WARN: go-secrets.yml not found — create go/k8s/secrets/go-secrets.yml with jwt-secret"
fi

echo "==> Applying monitoring RBAC..."
kubectl apply -f "$SCRIPT_DIR/monitoring/rbac/"

echo "==> Applying PVCs..."
kubectl apply -f "$SCRIPT_DIR/monitoring/pvc/"
kubectl apply -f "$REPO_DIR/java/k8s/volumes/"

echo "==> Applying ConfigMaps..."
kubectl apply -f "$SCRIPT_DIR/ai-services/configmaps/"
kubectl apply -f "$REPO_DIR/java/k8s/configmaps/"
kubectl apply -f "$SCRIPT_DIR/monitoring/configmaps/"
kubectl apply -f "$REPO_DIR/go/k8s/configmaps/"

echo "==> Deploying ai-services (Python)..."
kubectl apply -f "$SCRIPT_DIR/ai-services/deployments/"
kubectl apply -f "$SCRIPT_DIR/ai-services/services/"

echo "==> Waiting for Qdrant..."
kubectl wait --for=condition=available --timeout=120s deployment/qdrant -n ai-services

echo "==> Deploying java-tasks..."
kubectl apply -f "$REPO_DIR/java/k8s/deployments/"
kubectl apply -f "$REPO_DIR/java/k8s/services/"

echo "==> Waiting for java-tasks infrastructure..."
kubectl wait --for=condition=available --timeout=120s deployment/postgres -n java-tasks
kubectl wait --for=condition=available --timeout=120s deployment/mongodb -n java-tasks
kubectl wait --for=condition=available --timeout=120s deployment/redis -n java-tasks
kubectl wait --for=condition=available --timeout=120s deployment/rabbitmq -n java-tasks

echo "==> Running go-ecommerce migration jobs..."
kubectl apply -f "$REPO_DIR/go/k8s/jobs/"

echo "==> Deploying go-ecommerce services..."
kubectl apply -f "$REPO_DIR/go/k8s/deployments/"
kubectl apply -f "$REPO_DIR/go/k8s/services/"
kubectl apply -f "$REPO_DIR/go/k8s/hpa/"

echo "==> Deploying monitoring..."
kubectl apply -f "$SCRIPT_DIR/monitoring/deployments/"
kubectl apply -f "$SCRIPT_DIR/monitoring/services/"
kubectl apply -f "$SCRIPT_DIR/monitoring/daemonsets/"

echo "==> Applying Ingress resources..."
kubectl apply -f "$SCRIPT_DIR/ai-services/ingress.yml"
kubectl apply -f "$REPO_DIR/java/k8s/ingress.yml"
kubectl apply -f "$REPO_DIR/java/k8s/ingress-rabbitmq.yml"
kubectl apply -f "$SCRIPT_DIR/monitoring/ingress.yml"
kubectl apply -f "$REPO_DIR/go/k8s/ingress.yml"

echo "==> Waiting for all application services..."
kubectl wait --for=condition=available --timeout=180s deployment/ingestion -n ai-services
kubectl wait --for=condition=available --timeout=180s deployment/chat -n ai-services
kubectl wait --for=condition=available --timeout=180s deployment/debug -n ai-services
kubectl wait --for=condition=available --timeout=180s deployment/task-service -n java-tasks
kubectl wait --for=condition=available --timeout=180s deployment/activity-service -n java-tasks
kubectl wait --for=condition=available --timeout=180s deployment/notification-service -n java-tasks
kubectl wait --for=condition=available --timeout=180s deployment/gateway-service -n java-tasks
kubectl wait --for=condition=available --timeout=180s deployment/go-auth-service -n go-ecommerce
kubectl wait --for=condition=available --timeout=180s deployment/go-ecommerce-service -n go-ecommerce
kubectl wait --for=condition=available --timeout=180s deployment/go-ai-service -n go-ecommerce
kubectl wait --for=condition=available --timeout=120s deployment/prometheus -n monitoring
kubectl wait --for=condition=available --timeout=120s deployment/kube-state-metrics -n monitoring
kubectl wait --for=condition=available --timeout=120s deployment/grafana -n monitoring

echo ""
echo "==> All services deployed!"
echo ""
echo "    Namespaces:"
echo "      ai-services    — Python AI services + Qdrant"
echo "      java-tasks     — Java microservices + databases"
echo "      go-ecommerce   — Go auth + ecommerce + AI agent services"
echo "      monitoring     — Prometheus + Grafana"
echo ""
echo "    Next steps:"
echo "      1. Run 'minikube tunnel' in a separate terminal (requires sudo)"
echo "      2. Access services at http://localhost/"
echo ""
echo "    Endpoints (via Ingress):"
echo "      /ingestion/*    — Document ingestion API"
echo "      /chat/*         — RAG chat API"
echo "      /debug/*        — Debug assistant API"
echo "      /graphql        — Java GraphQL API"
echo "      /graphiql       — GraphQL IDE"
echo "      /auth/*         — OAuth authentication"
echo "      /go-auth/*      — Go auth API"
echo "      /go-api/*       — Go ecommerce API"
echo "      /ai-api/*       — Go AI agent API"
echo "      /grafana/       — Monitoring dashboards"
echo "      /rabbitmq/      — Message broker UI"
echo ""
echo "    Verify: kubectl get ingress --all-namespaces"
