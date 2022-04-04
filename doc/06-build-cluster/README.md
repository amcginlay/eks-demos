# Build EKS Cluster

Verify that your Cloud9 environment is currently assuming the `Role-EC2-EKSClusterAdmin` IAM role.
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
        externalDNS: true
        certManager: true

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

You can further validate your permissions by observing the pods initally deployed in the kube-system namespace.
```bash
kubectl -n kube-system get pods -o wide
```

[Return To Main Menu](/README.md)
