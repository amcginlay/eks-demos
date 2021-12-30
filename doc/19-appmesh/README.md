# (WIP) AWS App Mesh - because managing microservices at scale is hard

If you have **not** completed the earlier section on Services (Load Distribution) then you may not have an appropriate service manifest and corresponding service object in place.
If so, please return and complete the section named **"K8s ClusterIP Services"**.

This section also assumes you are familiar with the use of **Helm** for deployment of Kubernetes objects.

Install the [App Mesh Controller](https://aws.github.io/aws-app-mesh-controller-for-k8s/) as follows, ignoring any warnings.
```bash
eksctl create iamserviceaccount \
  --cluster ${EKS_CLUSTER_NAME} \
  --namespace kube-system \
  --name appmesh-controller \
  --attach-policy-arn arn:aws:iam::aws:policy/AWSCloudMapFullAccess,arn:aws:iam::aws:policy/AWSAppMeshFullAccess \
  --override-existing-serviceaccounts \
  --approve

helm repo add eks https://aws.github.io/eks-charts
helm -n kube-system upgrade -i appmesh-controller eks/appmesh-controller \
  --set region=${AWS_DEFAULT_REGION} \
  --set serviceAccount.create=false \
  --set serviceAccount.name=appmesh-controller \
  --set tracing.enabled=true \
  --set tracing.provider=x-ray
```

Verify that the App Mesh Controller is installed.
```bash
kubectl -n kube-system get deployment appmesh-controller
```

In the earlier section titled **Upgrade Your Deployment** we introduced version 2.0 of `echo-frontend` which supports the use of a backend app.
Using skills introduced in previous chapters, you will now rapidly build, containerize and register two versions of the `echo-backend` app now.
We aim to use App Mesh to demonstrate a Blue Green deployment model within which traffic can be dynamically shifted between two different versions of fully-scaled backends.
```bash
aws ecr delete-repository --repository-name echo-backend --force >/dev/null 2>&1
aws ecr create-repository \
  --repository-name echo-backend \
  --image-scanning-configuration scanOnPush=true \
  --query 'repository.repositoryUri' \
  --output text

aws ecr get-login-password | docker login --username AWS --password-stdin ${EKS_ECR_REGISTRY}

mkdir -p ~/environment/echo-backend/src/1.0/ 
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-backend/src/1.0/main.go \
     https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-backend/src/1.0/Dockerfile \
     --directory-prefix ~/environment/echo-backend/src/1.0/
docker build -t echo-backend:1.0 ~/environment/echo-backend/src/1.0/
docker tag echo-backend:1.0 ${EKS_ECR_REGISTRY}/echo-backend:1.0
docker push ${EKS_ECR_REGISTRY}/echo-backend:1.0

mkdir -p ~/environment/echo-backend/src/2.0/ 
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-backend/src/2.0/main.go \
     https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-backend/src/2.0/Dockerfile \
     --directory-prefix ~/environment/echo-backend/src/2.0/
docker build -t echo-backend:2.0 ~/environment/echo-backend/src/2.0/
docker tag echo-backend:2.0 ${EKS_ECR_REGISTRY}/echo-backend:2.0
docker push ${EKS_ECR_REGISTRY}/echo-backend:2.0
```

 .... TODO ....

[Return To Main Menu](/README.md)
