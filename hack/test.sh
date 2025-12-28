#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# 사용법 체크: hack 디렉터리 안에서 직접 실행하지 말 것
if [[ "$(pwd)" == "${ROOT}/hack" ]]; then
  echo "================================================================"
  echo "  [operator-slo-lab] 잘못된 실행 위치입니다."
  echo
  echo "  이 스크립트는 리포지토리 루트에서 실행해야 합니다."
  echo "    예) \$ cd ${ROOT}"
  echo "        \$ ./hack/test.sh"
  echo
  echo "  현재 디렉터리: $(pwd)"
  echo "================================================================"
  exit 1
fi

# Pick the newest envtest assets directory under bin/k8s (e.g., 1.34.1-linux-amd64)
LATEST="$(ls -1 "${ROOT}/bin/k8s" | sort -V | tail -1)"
export KUBEBUILDER_ASSETS="${ROOT}/bin/k8s/${LATEST}"

echo "Using KUBEBUILDER_ASSETS=${KUBEBUILDER_ASSETS}"

# 이중안전장치
cd "${ROOT}"
go test ./...
