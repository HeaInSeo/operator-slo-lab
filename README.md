# operator-slo-lab
###
- High-Performance/Complex Lifecycle Operator를 위한 SLI/SLO 프레임워크

### 시작하기

```bash

# 터미널 A 루트에서
./hack/dev-start.sh # kind 클러스터 확인 및 정검

# 새터미널에서 루트에서         
make install        # CRD 등록
make run            # manager 실행 (로그 관찰 창으로 두기)

# 터미널 A 루트에서 (namespace 잊지말자. 자꾸 잊어버림. 기억용으로 남김.)
kubectl create namespace slo-lab
kubectl get ns # 확인
kubectl apply -n slo-lab -f config/samples/lab_v1alpha1_slojob.yaml
kubectl get slojob -n slo-lab

#kubectl annotate slojob slojob-sample \
#  -n slo-lab \
#  test/start-time="$(date --rfc3339=ns)" \
#  --overwrite

# curl -s "$METRICS" | grep e2e_convergence | head

# dev-start.sh 에서 METRICS_DEFAULT="http://localhost:8080/metrics" 이렇게 설정하지만, 만약 다른 터미널에서 사용하게 되면 이렇게 설정하라고 남겨둠.
METRICS=${METRICS:-http://localhost:8080/metrics}

# metrics server 살아있고 열려있는지 확인.
curl -sf "$METRICS" >/dev/null \
  && echo "[OK] metrics endpoint alive: $METRICS" \
  || echo "[FAIL] cannot reach metrics endpoint: $METRICS"

```

### TODO
- kubectl 관련 shellscript 로 정리해놓자. ns 자꾸 잊어버리는데 이거 잊지말고 yaml 에 넣지 말고 kubectl 로 ns 에 넣어주는 방식으로 간다.  

### fork 할 때 주의사항
- kubectl 관련해서 wrapping 한 스크립트를 쓰기 때문에 kubectl 설치해서 사용하는 개념이 아니다.
- 아래와 같이 작성해서 사용하고 있고, /usr/local/bin 에 넣어두어서 전역적으로 사용할 수 있게 만들었다.

```bash
  GNU nano 6.2                                                                                       kubectl                                                                                                
#!/usr/bin/env bash
set -euo pipefail

# Wrapper script to run kubectl inside a container with proper kubeconfig mount

# Directory containing host kubeconfig
KUBEDIR="${HOME}/.kube"

# Ensure .kube directory exists
mkdir -p "$KUBEDIR"

# Warn if kubeconfig file is missing
if [ ! -f "$KUBEDIR/config" ]; then
  echo "WARNING: kubeconfig not found at $KUBEDIR/config."
  echo "Proceeding without existing or valid config."
fi

# Pre-check: script identification
# v1.0.2  kubectl-tree 마운트 시킴. 
echo "▶ $(basename "$0") v1.0.2 - wrapper script for kubectl"

# [추가됨]-플러그인 경로 설정
PLUGIN_DIR="${HOME}/.kube/plugins"

# Run kubectl in a container, passing all arguments ($@) directly to kubectl

exec docker run --rm -i \
  --network host \
  --user root \
  -v "${KUBEDIR}:/root/.kube" \
  -v "/opt/.kube:/opt/.kube" \
  -v "$(pwd):/workdir" \
  -v "${PLUGIN_DIR}/kubectl-tree:/usr/local/bin/kubectl-tree" \
  -e KUBECONFIG=/root/.kube/config \
  -w /workdir \
  bitnami/kubectl:latest \
  "$@"

```

- 이렇게 kubectl 을 사용하기 때문에, ./hack/dev-start.sh 내용중에, 
- 해당 shell script 에서 kubectl 을 실행할때,  echo "▶ $(basename "$0") v1.0.2 - wrapper script for kubectl" 이렇게 출력되는 것을 지울 필요가 있어,
- 아래와 같이 작성하였다. 따라서, 직접 kubectl 을 사용할 경우는 이 shell script(dev-start.sh) 를 수정해서 사용하여야 한다.

```bash
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
  log "  ⚠️  current-context 출력 (정제 전):"
  echo "      ${ctx_raw}"
  log "  ⚠️  current-context 출력 (정제 후):"
  echo "      ${ctx_clean}"
  log "    → 'kind-slo-lab' 로 전환 권장:"
  echo "      kubectl config use-context kind-slo-lab"
fi
```