# operator-slo-lab
###
- High-Performance/Complex Lifecycle Operator를 위한 SLI/SLO 프레임워크

### 시작하기
- hack 폴더에서 dev-start.sh 을 통해서 시작 및 정검을 할 수 있다. 
- 정상적으로 시작됨다면 다른 터미널에서 make run 으로 시작할 수 있다. 물론 사전에 설정 및 설치관련은 완료 되어 있어야 한다. 이것은 노션을 참고하면 된다. (모르면 질문)


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