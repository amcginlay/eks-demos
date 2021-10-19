# Configure Client Tools

Install AWS CLI v2, eksctl, kubectl, AWS Session Manager plugin, jq, helm, tree, siege and gradle.
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
sudo yum install -y session-manager-plugin.rpm jq tree siege
curl -sSL https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
curl -s "https://get.sdkman.io" | bash
source "$HOME/.sdkman/bin/sdkman-init.sh"
sdk install gradle
rm -r ./aws/ ./awscliv2.zip session-manager-plugin.rpm

# finally, install the kubectl neat add-on (https://krew.sigs.k8s.io/docs/user-guide/setup/install/ | https://github.com/itaysk/kubectl-neat)
(
  set -x; cd "$(mktemp -d)" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/krew.tar.gz" &&
  tar zxvf krew.tar.gz &&
  KREW=./krew-"$(uname | tr '[:upper:]' '[:lower:]')_$(uname -m | sed -e 's/x86_64/amd64/' -e 's/arm.*$/arm/')" &&
  "$KREW" install krew
)
echo 'export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
kubectl krew install neat
```

Verify the installs worked.
```bash
which aws eksctl kubectl session-manager-plugin jq tree helm siege gradle krew
```

Each Cloud9 instance has the [Docker](https://en.wikipedia.org/wiki/Docker_(software)) daemon installed with a set of images pre-loaded. Remove them as they are not required.
```bash
for i in $(docker ps -q); do docker kill $i; done
docker system prune --all --force
```

[Return To Main Menu](/README.md)
