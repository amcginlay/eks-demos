# Build EKS Cluster

**IMPORTANT MANUAL STEPS** - At this point, c9 is using "AWS managed temporary credentials" and we are NOT currently assuming the Role-EC2-EKSClusterAdmin. You must perform the following step manually:

* **go to c9 IDE Preferences**
* **-> AWS Settings**
* **-> switch OFF "AWS managed temporary credentials"**

Verify we are now using the Role-EC2-EKSClusterAdmin IAM role.
```bash
aws sts get-caller-identity
```

Set up a KMS customer managed key to encrypt secrets, as per: https://aws.amazon.com/blogs/containers/using-eks-encryption-provider-support-for-defense-in-depth/
```bash
key_metadata=($(aws kms create-key --query KeyMetadata.[KeyId,Arn] --output text)) # [0]=KeyId [1]=Arn
aws kms create-alias --alias-name alias/cmk-eks-${cluster_name}-$(cut -c-8 <<< ${key_metadata[0]}) --target-key-id ${key_metadata[1]}
```

Create a manifest describing the EKS cluster with a managed node group and fargate profile (NOTE "eksctl create cluster" will also update ~/.kube/config)
```bash
cat > ~/environment/${cluster_name}-cluster-config.yaml << EOF
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: ${cluster_name}
  region: ${AWS_DEFAULT_REGION}
  version: "${k8s_version}"
availabilityZones: ["${AWS_DEFAULT_REGION}a", "${AWS_DEFAULT_REGION}b", "${AWS_DEFAULT_REGION}c"]
secretsEncryption:
  keyARN: ${key_metadata[1]}
managedNodeGroups:
  - name: ng-${cluster_name}
    availabilityZones: ["${AWS_DEFAULT_REGION}a", "${AWS_DEFAULT_REGION}b", "${AWS_DEFAULT_REGION}c"]
    instanceType: t3.small
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
  - name: fp-${cluster_name}
    selectors:
      - namespace: serverless
EOF
```

Build the EKS cluster from the manifest (~20 mins)
```bash
eksctl create cluster -f ~/environment/${cluster_name}-cluster-config.yaml 
```

Check the Cloud9 environment can connect to the k8s cluster and display the TWO nodes in the managed node group.
```bash
kubectl get nodes -o wide
```

[Return To Main Menu](../README.md)
