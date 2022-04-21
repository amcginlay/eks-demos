# Build EKS Cluster

Verify that your Cloud9 environment is currently assuming the `Role-EC2-EKSClusterAdmin` IAM role.
```bash
aws sts get-caller-identity
```

Set up a KMS customer managed key to encrypt secrets, as per: https://aws.amazon.com/blogs/containers/using-eks-encryption-provider-support-for-defense-in-depth/
```bash
key_metadata=($(aws kms create-key --query KeyMetadata.[KeyId,Arn] --output text)) # [0]=KeyId [1]=Arn
aws kms create-alias --alias-name alias/${C9_PROJECT}-$(cut -c-8 <<< ${key_metadata[0]}) --target-key-id ${key_metadata[1]}
```

Create a manifest describing the EKS cluster with a managed node group (using spot instances) alongside a fargate profile.
```bash
cat > ~/environment/${C9_PROJECT}-cluster-config.yaml << EOF
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: ${C9_PROJECT}
  region: ${AWS_DEFAULT_REGION}
  version: "${EKS_K8S_VERSION}"
availabilityZones: ["${AWS_DEFAULT_REGION}a", "${AWS_DEFAULT_REGION}b", "${AWS_DEFAULT_REGION}c"]
secretsEncryption:
  keyARN: ${key_metadata[1]}
iam:
  withOIDC: true
addons: # the usual suspects - accept defaults, formalize existence (see Console : Cluster -> Configuration -> Add-ons)
  - name: coredns
  - name: kube-proxy
  - name: vpc-cni
cloudWatch: # comment out as necessary
  clusterLogging:
    enableTypes:
      - "api"
      - "audit"
      - "authenticator"
      - "controllerManager"
      - "scheduler"

managedNodeGroups:
  - name: mng
    availabilityZones: ["${AWS_DEFAULT_REGION}a", "${AWS_DEFAULT_REGION}b", "${AWS_DEFAULT_REGION}c"]
    instanceTypes: ["t3.small","t3a.small"]
    privateNetworking: true
    spot: true
    desiredCapacity: 2
    maxSize: 6
    iam:
      attachPolicyARNs:      
        - arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy
        - arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy
        - arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly
        - arn:aws:iam::aws:policy/AWSCertificateManagerPrivateCAFullAccess
        - arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore
      withAddonPolicies:
        autoScaler: true
        appMesh: true
        albIngress: true
        xRay: true
        cloudWatch: true
        externalDNS: true
        certManager: true

# we do not want to concern ourselves with self managed nodes, but here's how eksctl handles them
# nodeGroups:
#   - name: ng
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
  - name: fp
    selectors:
      - namespace: serverless
EOF
```

Build the EKS cluster from the manifest (~20 mins). NOTE this will also update `~/.kube/config`
```bash
eksctl create cluster -f ~/environment/${C9_PROJECT}-cluster-config.yaml 
```

Check the Cloud9 environment can connect to the k8s cluster and display the TWO nodes in the managed node group.
```bash
kubectl get nodes -o wide
```

You can further validate your permissions by observing the pods initally deployed in the kube-system namespace.
```bash
kubectl -n kube-system get pods -o wide
```

## Configure SSM access (optional)

`eksctl` has already put your worker nodes into private subnets and the SSH port (22) is closed.
This is good practice but what if you still require occasional remote access to these EC2 instances for diagnostic purposes?

Here's the scripted equivalent of [this](https://aws.amazon.com/premiumsupport/knowledge-center/ec2-systems-manager-vpc-endpoints/) knowledge base article which opens up your private worker nodes to Systems Manager (SSM) Session Manager via VPC Endpoints.
Note that the IAM role for your worker nodes already has the `AmazonSSMManagedInstanceCore` policy attached which is part of this solution (see your cluster config YAML file above).

```bash
CLUSTER_NAME=${C9_PROJECT}
vpc=$(aws eks describe-cluster --name ${CLUSTER_NAME} --query cluster.resourcesVpcConfig.vpcId --output text)
subnets=($(aws ec2 describe-subnets | jq --arg CLUSTER_NAME "$CLUSTER_NAME" '.Subnets[] | select(contains({Tags: [{Key: "Name"}, {Value: $CLUSTER_NAME}]}) and contains({Tags: [{Key: "Name"}, {Value: "Private"}]})) | .SubnetId' --raw-output))
sg=$(aws ec2 create-security-group --group-name allow-https-${CLUSTER_NAME} --description allow-https-${CLUSTER_NAME} --vpc-id ${vpc} --query GroupId --output text)
aws ec2 authorize-security-group-ingress --group-id ${sg} --protocol tcp --port 443 --cidr 0.0.0.0/0
aws ec2 create-vpc-endpoint --vpc-id ${vpc} --service-name com.amazonaws.us-west-2.ssm --vpc-endpoint-type Interface --subnet-ids ${subnets[*]} --security-group-ids ${sg}
aws ec2 create-vpc-endpoint --vpc-id ${vpc} --service-name com.amazonaws.us-west-2.ssmmessages --vpc-endpoint-type Interface --subnet-ids ${subnets[*]} --security-group-ids ${sg}
aws ec2 create-vpc-endpoint --vpc-id ${vpc} --service-name com.amazonaws.us-west-2.ec2messages --vpc-endpoint-type Interface --subnet-ids ${subnets[*]} --security-group-ids ${sg}
```

Since you already have the SSM CLI plugin installed, you may now connect to your worker nodes using the following.
```bash
aws ssm start-session --target <EC2_INSTANCE_ID>
```

Next: [Main Menu](/README.md) | [Configure Local Machine Access](../07-local-access/README.md)
