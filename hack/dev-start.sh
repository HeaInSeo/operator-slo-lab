#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

log() {
  echo "[$(date +'%H:%M:%S')] $*"
}

# 권장 실행 위치 안내 (루트에서 ./hack/dev-start.sh)
if [[ "$(pwd)" == "${ROOT}/hack" ]]; then
  echo "================================================================"
  echo "  [operator-slo-lab] 권장 실행 위치는 리포지토리 루트입니다."
  echo
  echo "    예) \$ cd ${ROOT}"
  echo "        \$ ./hack/dev-start.sh"
  echo
  echo "  현재 디렉터리: $(pwd)"
  echo "================================================================"
  # 필요하면 여기서 exit 1 해도 됨. 지금은 안내만.
fi

log "1) kind 클러스터(slo-lab) 존재 여부 확인"
if ! kind get clusters | grep -q '^slo-lab$'; then
  log "  ✗ kind 클러스터 'slo-lab' 을 찾지 못했습니다."
  log "    → 먼저 아래 명령으로 클러스터를 만들고 다시 실행해주세요:"
  echo "      kind create cluster --name slo-lab"
  exit 1
else
  log "  ✓ kind 클러스터 'slo-lab' 감지"
fi

log "2) kubectl current-context 확인"

# kubectl 출력 (wrapper 포함 가능)
ctx_raw="$(kubectl config current-context 2>/dev/null || echo '<none>')"

# 'wrapper script for kubectl' 이 들어간 줄 제거 + 빈 줄 제거
ctx_clean="$(
  printf '%s\n' "${ctx_raw}" \
    | sed '/wrapper script for kubectl/d' \
    | sed '/^$/d'
)"

if [[ "${ctx_clean}" == "kind-slo-lab" ]]; then
  log "  ✓ current-context = kind-slo-lab"
else
  log "  ⚠️ current-context 출력 (정제 전):"
  echo "      ${ctx_raw}"
  log "  ⚠️ current-context 출력 (정제 후):"
  echo "      ${ctx_clean}"
  log "    → 'kind-slo-lab' 로 전환 권장:"
  echo "      kubectl config use-context kind-slo-lab"
fi

log "3) 노드 상태 확인 (kubectl get nodes)"
kubectl get nodes

log "4) 테스트 실행 (./hack/test.sh)"
# 단일 진입점으로 통합하기 위해 주석 처리
#"${ROOT}/hack/test.sh"
make -C "${ROOT}" test

log "5) METRICS 기본값 안내"
METRICS="http://localhost:8080/metrics"
log "  기본 METRICS URL: ${METRICS}"
echo
#echo "  ※ manager(make run)가 이미 떠 있다면, 이 터미널(또는 다른 터미널)에서 예를 들어:"
echo
echo "    METRICS=\${METRICS:-${METRICS}}"
#echo "    curl -s \"\$METRICS\" | grep e2e_convergence_time | head"
#echo "    curl -s \"\$METRICS\" | grep workqueue_ | head"
echo
log "=== dev-start 정검 완료. 이제 다음 순서로 진행하면 됩니다 ==="
echo "  Readme.md 의 '시작하기' 섹션 참고"
#echo "  1) 새 터미널을 열어서, 리포지토리 루트에서:"
#echo "       make run"
#echo "     (이 터미널은 manager 로그만 띄워두고 두는 것을 추천합니다.)"
#echo
#echo "  2) dev-start.sh 를 실행한 현재 터미널에서 샘플 적용 및 metrics 확인:"
#echo "       kubectl apply -f config/samples"
#echo "       METRICS=\${METRICS:-${METRICS}}"
#echo "       curl -s \"\$METRICS\" | grep e2e_convergence_time | head"
#echo "       curl -s \"\$METRICS\" | grep workqueue_ | head"
#echo
