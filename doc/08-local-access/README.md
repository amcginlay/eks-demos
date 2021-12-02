# Configure Local Machine Access

This module is optional.
Subsequent labs will only assume connectivity via Cloud9.

The EC2 instance hosting your Cloud9 environment is assuming the `Role-EC2-EKSClusterAdmin` role you created earlier.
As the cluster creator, this role is **implcitly** a member of the k8s RBAC group named `system:masters` which represents the cluster administrators.
As a result, this IAM role currently represents the one and only trusted administrator of the cluster.
If you wish to include further administrator identities you can now introduce their these to the cluster.

A common expectation is to be able to connect to the cluster from a local machine.
We assume this local machine already has up to date versions of the `aws` and `kubectl` tools installed.
We also assume this local machine is [configured to access the AWS account](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) and can provide a Bash shell.

As this cluster is unknown to your **local machine** it will need an appropriately configured kubeconfig file installed at `~/.kube/config`.
This will get your local machine authenticated (but not yet authorized) with the cluster.
```bash
cluster=dev
aws eks update-kubeconfig --name ${cluster}
```

The following command will confirm the **unauthorized** status of your **local machine**.
```bash
kubectl get nodes
```

To address this, first run the following Bash shell commands on your **local machine** to identify the ARN of the currently configured IAM principal.
```bash
arn=$(aws sts get-caller-identity --query Arn --output text)
new_admin_arn=${arn}
if grep -q assumed-role <<< ${arn}; then
  new_admin_arn=$((sed "s/:sts:/:iam:/g" | sed "s/assumed-//g" | rev | cut -d/ -f2- | rev) <<< ${arn})
fi
echo ${new_admin_arn}
```

Then, in the **Cloud9 terminal**, run the following `eksctl` command, ensuring that you first update the `<NEW_ADMIN_ARN>` placeholder as appropriate.
This will **explicitly** introduce the new administrator to the cluster.
```bash
new_admin_arn=<NEW_ADMIN_ARN>
eksctl create iamidentitymapping \
  --cluster dev \
  --group system:masters \
  --arn ${new_admin_arn} \
  --username $((rev | cut -d/ -f1 | rev) <<< ${new_admin_arn})
```

Back on your **local machine**, run this command to confirm successful authorization.
```bash
kubectl get nodes
```

Behind the scenes, the call to `eksctl create iamidentitymapping` updated the `aws-auth` configmap which acts as a bridge between authentication (AWS IAM) and authorization (Kubernetes RBAC).
This k8s configmap, which is unique to EKS clusters, resides in the `kube-system` namespace.
Only the cluster creator is initially authorized to interact with the cluster.
Appropriate entries in the `aws-auth` configmap are required before the cluster will acknowledge any others.
You can view the configmap from any authorized client device using the following.
```bash
kubectl -n kube-system get configmap aws-auth -o yaml
```

You may, of course, add entries to the `aws-auth` configmap manually but `eksctl create iamidentitymapping` is the safer option.

[Return To Main Menu](/README.md)
