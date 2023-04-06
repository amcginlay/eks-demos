# Configure Client Tools

Install AWS CLI v2, eksctl, kubectl, AWS Session Manager plugin, dive, jq, helm, tree, siege, gradle and krew.
```bash
sudo mv /usr/local/bin/aws /usr/local/bin/aws.old
sudo mv /usr/bin/aws /usr/bin/aws.old
curl --silent "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
sudo mv /tmp/eksctl /usr/local/bin
curl -LO https://storage.googleapis.com/kubernetes-release/release/v${EKS_K8S_VERSION}.0/bin/linux/amd64/kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
curl --silent "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/linux_64bit/session-manager-plugin.rpm" -o "session-manager-plugin.rpm"
dive_latest=$(curl -s https://api.github.com/repos/wagoodman/dive/releases/latest | grep -oP '"tag_name": "\K(.*)(?=")')
curl --silent --location https://github.com/wagoodman/dive/releases/download/${dive_latest}/dive_$(cut -c 2- <<< ${dive_latest})_linux_amd64.rpm -o dive.rpm 
sudo yum install -y session-manager-plugin.rpm dive.rpm jq tree siege
curl -sSL https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
rm -r ./aws/ ./awscliv2.zip session-manager-plugin.rpm dive.rpm
# finally, install the kubectl neat add-on (https://krew.sigs.k8s.io/docs/user-guide/setup/install/ | https://github.com/itaysk/kubectl-neat)
(
  set -x; cd "$(mktemp -d)" &&
  OS="$(uname | tr '[:upper:]' '[:lower:]')" &&
  ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')" &&
  KREW="krew-${OS}_${ARCH}" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" &&
  tar zxvf "${KREW}.tar.gz" &&
  ./"${KREW}" install krew
)
echo 'export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
kubectl krew install neat get-all
```

Verify the installs worked.
```bash
which aws eksctl kubectl session-manager-plugin dive jq tree helm siege kubectl-neat
```

Configure [kubectl autocomplete](https://kubernetes.io/docs/tasks/tools/included/optional-kubectl-configs-bash-linux/).
```bash
cat > ~/.kubectl-ac << EOF
source <(kubectl completion bash)
alias k=kubectl
complete -F __start_kubectl k
EOF
echo "source ~/.kubectl-ac" >> ~/.bashrc
source ~/.kubectl-ac
```

Next: [Main Menu](/README.md) | [Build EKS Cluster](../06-build-cluster/README.md)
