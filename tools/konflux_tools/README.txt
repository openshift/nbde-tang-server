Access Konflux cluster externally:
----------------------------------
# 1 - Install krew (as kubectl cluster handler)
(
  set -x; cd "$(mktemp -d)" &&
  OS="$(uname | tr '[:upper:]' '[:lower:]')" &&
  ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" &&
  KREW="krew-${OS}_${ARCH}" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" &&
  tar zxvf "${KREW}.tar.gz" &&
  ./"${KREW}" install krew
)
# 2 - Set krew path
export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"

# 3 - Install OIDC plugin
kubectl krew install oidc-login

# 4 - Set configuration for your tenant. In case of the konflux-sec-eng-spec:
apiVersion: v1
clusters:
- cluster:
    server: https://api-toolchain-host-operator.apps.stone-prd-host1.wdlc.p1.openshiftapps.com/workspaces/konflux-sec-eng-spec
  name: appstudio
contexts:
- context:
    cluster: appstudio
    namespace: konflux-sec-eng-spec-tenant
    user: oidc
  name: appstudio
current-context: appstudio
kind: Config
preferences: {}
users:
- name: oidc
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - oidc-login
      - get-token
      - --oidc-issuer-url=https://sso.redhat.com/auth/realms/redhat-external
      - --oidc-client-id=rhoas-cli-prod
      command: kubectl
      env: null
      provideClusterInfo: false

# 5 - Save previous file as konflux-cluster.txt and apply configuration in this directory as a kubeconfig file:
./go.sh
