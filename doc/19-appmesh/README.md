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

We aim to use App Mesh to demonstrate a [Blue-green deployment](https://en.wikipedia.org/wiki/Blue-green_deployment) whereby traffic between deployments (i.e. microservices) can be dynamically shifted from one target to another without the need to re-deploy any existing workloads.
In the earlier section titled **Upgrade Your Deployment** we introduced version 2.0 of `echo-frontend` which supports the use of a backend app.
Using skills introduced in previous chapters you will now rapidly build, containerize, register and deploy two versions of the `echo-backend` app.
```bash
aws ecr delete-repository --repository-name echo-backend --force >/dev/null 2>&1
aws ecr create-repository \
  --repository-name echo-backend \
  --image-scanning-configuration scanOnPush=true \
  --query 'repository.repositoryUri' \
  --output text

aws ecr get-login-password | docker login --username AWS --password-stdin ${EKS_ECR_REGISTRY}

wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-backend/templates/echo-backend-deployment.yaml \
  -O ~/environment/echo-backend/templates/echo-backend-deployment.yaml
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-backend/templates/echo-backend-service.yaml \
  -O ~/environment/echo-backend/templates/echo-backend-service.yaml

declare -A versions=()
versions[blue]=1.0
versions[green]=2.0
for color in blue green; do
  version=${versions[${color}]}
  mkdir -p ~/environment/echo-backend/src/${version}/ 
  wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-backend/src/${version}/main.go \
    -O ~/environment/echo-backend/src/${version}/main.go
  wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-backend/src/${version}/Dockerfile \
    -O ~/environment/echo-backend/templates/src/${version}/Dockerfile
  docker build -t echo-backend:${version} ~/environment/echo-backend/src/${version}/
  docker tag echo-backend:${version} ${EKS_ECR_REGISTRY}/echo-backend:${version}
  docker push ${EKS_ECR_REGISTRY}/echo-backend:${version}

  cat ~/environment/echo-backend/templates/echo-backend-deployment.yaml \
      <(echo ---) \
      ~/environment/echo-backend/templates/echo-backend-service.yaml | \
      sed "s/{{ .Values.registry }}/${EKS_ECR_REGISTRY}/g" | \
      sed "s/{{ .Values.color }}/${color}/g" | \
      sed "s/{{ .Values.version }}/${version}/g" | \
      kubectl -n demos apply -f -
done
```

 .... TODO ....

[Return To Main Menu](/README.md)
