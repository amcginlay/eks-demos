# Configure Client Tools

Install AWS CLI v2, eksctl, kubectl, AWS Session Manager plugin, jq, helm, tree and siege.
```bash
sudo mv /usr/local/bin/aws /usr/local/bin/aws.old
sudo mv /usr/bin/aws /usr/bin/aws.old
curl --silent "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
sudo mv /tmp/eksctl /usr/local/bin
curl -LO https://storage.googleapis.com/kubernetes-release/release/v${k8s_version}.0/bin/linux/amd64/kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
curl --silent "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/linux_64bit/session-manager-plugin.rpm" -o "session-manager-plugin.rpm"
sudo yum install -y session-manager-plugin.rpm jq tree siege
curl -sSL https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
rm -r ./aws/ ./awscliv2.zip session-manager-plugin.rpm

# ... plus a custom script to simplify remote calls to the EKS worker nodes via SSM
cat > ./ssm-exec << EOF
#!/bin/bash
worker_node_instance_id=\$1
command=\$2
command_id=\$(aws ssm send-command --instance-ids \${worker_node_instance_id} --document-name "AWS-RunShellScript" --parameters commands="\${command}" --output text --query Command.CommandId)
aws ssm wait command-executed --instance-id \${worker_node_instance_id} --command-id \${command_id}
aws ssm list-command-invocations --instance-id \${worker_node_instance_id} --command-id \${command_id} --details --output text --query CommandInvocations[0].CommandPlugins[0].Output
EOF
chmod +x ./ssm-exec
sudo mv ./ssm-exec /usr/local/bin/ssm-exec

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
which aws eksctl kubectl session-manager-plugin jq tree helm siege ssm-exec
```

[Return To Main Menu](../README.md)
