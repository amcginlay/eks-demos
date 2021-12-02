# Build EKS Cluster

Verify that your Cloud9 environment is currently assuming the Role-EC2-EKSClusterAdmin IAM role.
```bash
aws sts get-caller-identity
```

Set up a KMS customer managed key to encrypt secrets, as per: https://aws.amazon.com/blogs/containers/using-eks-encryption-provider-support-for-defense-in-depth/
```bash
key_metadata=($(aws kms create-key --query KeyMetadata.[KeyId,Arn] --output text)) # [0]=KeyId [1]=Arn
aws kms create-alias --alias-name alias/cmk-eks-${EKS_CLUSTER_NAME}-$(cut -c-8 <<< ${key_metadata[0]}) --target-key-id ${key_metadata[1]}
```

Create a manifest describing the EKS cluster with a managed node group (using spot instances) alongside a fargate profile.
```bash
cat > ~/environment/${EKS_CLUSTER_NAME}-cluster-config.yaml << EOF
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: ${EKS_CLUSTER_NAME}
  region: ${AWS_DEFAULT_REGION}
  version: "${EKS_K8S_VERSION}"
availabilityZones: ["${AWS_DEFAULT_REGION}a", "${AWS_DEFAULT_REGION}b", "${AWS_DEFAULT_REGION}c"]
secretsEncryption:
  keyARN: ${key_metadata[1]}
iam:
  withOIDC: true
  
managedNodeGroups:
  - name: mng-${EKS_CLUSTER_NAME}
    availabilityZones: ["${AWS_DEFAULT_REGION}a", "${AWS_DEFAULT_REGION}b", "${AWS_DEFAULT_REGION}c"]
    instanceTypes: ["t3.small","t3a.small"]
    privateNetworking: true
    spot: true
    desiredCapacity: 2
    maxSize: 6
    iam:
      withAddonPolicies:
        autoScaler: true
        appMesh: true
        albIngress: true
        xRay: true
        cloudWatch: true

# we do not want to concern ourselves with self managed nodes, but here's how eksctl handles them
# nodeGroups:
#   - name: ng-${EKS_CLUSTER_NAME}
#     availabilityZones: ["us-west-2a", "us-west-2b", "us-west-2c"]
#     instanceType: "t3.small"
#     privateNetworking: true
#     desiredCapacity: 1
#     maxSize: 1
#     taints:
#       - key: node-type
#         value: self-managed
#         effect: NoSchedule

fargateProfiles:
  - name: fp-${EKS_CLUSTER_NAME}
    selectors:
      - namespace: serverless
EOF
```

Build the EKS cluster from the manifest (~20 mins). NOTE this will also update `~/.kube/config`
```bash
eksctl create cluster -f ~/environment/${EKS_CLUSTER_NAME}-cluster-config.yaml 
```

Check the Cloud9 environment can connect to the k8s cluster and display the TWO nodes in the managed node group.
```bash
kubectl get nodes -o wide
```

We can further validate our permissions by observing the pods initally deployed in the kube-system namespace.
```bash
kubectl -n kube-system get pods -o wide
```

The EC2 instance hosting your Cloud9 environment is assuming the `Role-EC2-EKSClusterAdmin` role you created earlier.
As the cluster creator, this role is implcitly a member of the k8s RBAC group named `system:masters` which represents the cluster administrators.
As a result, this role is currently the **only** trusted administrator of the cluster.
If you wish to include further administrator identities you can now introduce their these to the cluster.

A common expectation is to be able to connect to the cluster from a local machine.
We assume this local machine already has up to date versions of the `aws` and `kubectl` tools installed.
We also assume this local machine is [configured to access the AWS account](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) and can provide a Bash shell.

As this cluster is unknown to your local machine it will need an appropriately configured kubeconfig file installed at `~/.kube/config`.
This will get your local machine authenticated (but not yet authorized) with the cluster.
```bash
cluster=dev
aws eks update-kubeconfig --name ${cluster}
```

The following command will confirm the **unauthorized** status of your local machine.
```bash
kubectl get nodes
```

To address this, first run the following Bash shell commands on your local machine to identify the ARN of the currently configured IAM principal.
```bash
arn=$(aws sts get-caller-identity --query Arn --output text)
new_admin_arn=${arn}
if grep -q assumed-role <<< ${arn}; then
  new_admin_arn=$((sed "s/:sts:/:iam:/g" | sed "s/assumed-//g" | rev | cut -d/ -f2- | rev) <<< ${arn})
fi
echo ${new_admin_arn}
```

Then, in the Cloud9 terminal, run the following `eksctl` command, ensuring that you first update the `<NEW_ADMIN_ARN>` placeholder as appropriate.
This will introduce the new administrator to the cluster.
```bash
new_admin_arn=<NEW_ADMIN_ARN>
eksctl create iamidentitymapping \
  --cluster dev \
  --group system:masters \
  --arn ${new_admin_arn} \
  --username $((rev | cut -d/ -f1 | rev) <<< ${new_admin_arn})
```

Back on your local machine, run this command to confirm successful authorization.
```bash
kubectl get nodes
```

Behind the scenes, the call to `eksctl create iamidentitymapping` updated the `aws-auth` configmap which acts as a bridge between **authentication** (AWS IAM) and **authorization** (Kubernetes RBAC).
This configmap resides in the `kube-system` namespace.
Everyone, except the cluster creator, is initally unauthorized to interact with the cluster until an appropriate entry is added to `aws-auth`.
You can view the configmap at any time using the following.
```bash
kubectl -n kube-system get configmap aws-auth -o yaml | kubectl neat
```

You may, of course, add entries to the `aws-auth` configmap manually but `eksctl create iamidentitymapping` is the safer option.

[Return To Main Menu](/README.md)
