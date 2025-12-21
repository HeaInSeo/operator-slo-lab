#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Pick the newest envtest assets directory under bin/k8s (e.g., 1.34.1-linux-amd64)
LATEST="$(ls -1 "${ROOT}/bin/k8s" | sort -V | tail -1)"
export KUBEBUILDER_ASSETS="${ROOT}/bin/k8s/${LATEST}"

echo "Using KUBEBUILDER_ASSETS=${KUBEBUILDER_ASSETS}"
go test ./...
