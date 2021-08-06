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

Create a manifest describing the EKS cluster with a managed node group and fargate profile (NOTE "eksctl create cluster" will also update ~/.kube/config)
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
managedNodeGroups:
  - name: ng-${EKS_CLUSTER_NAME}
    availabilityZones: ["${AWS_DEFAULT_REGION}a", "${AWS_DEFAULT_REGION}b", "${AWS_DEFAULT_REGION}c"]
    spot: true
    instanceTypes: ["t3.small","t3a.small"]
    desiredCapacity: 2
    maxSize: 6
    ssh:
      enableSsm: true
    iam:
      withAddonPolicies:
        autoScaler: true
        appMesh: true
        albIngress: true
        xRay: true
        cloudWatch: true
fargateProfiles:
  - name: fp-${EKS_CLUSTER_NAME}
    selectors:
      - namespace: serverless
EOF
```

Build the EKS cluster from the manifest (~20 mins)
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

[Return To Main Menu](/README.md)
