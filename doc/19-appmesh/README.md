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

cat > ~/environment/echo-backend/Chart.yaml << EOF
apiVersion: v2
name: echo-backend
version: 1.0.0
EOF

declare -A versions=()
versions[blue]=1.0
versions[green]=2.0
for color in blue green; do
  version=${versions[${color}]}
  mkdir -p ~/environment/echo-backend/src/${version}/
  wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-backend/src/${version}/main.go \
    -O ~/environment/echo-backend/src/${version}/main.go
  wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-backend/src/${version}/Dockerfile \
    -O ~/environment/echo-backend/src/${version}/Dockerfile
  docker build -t echo-backend:${version} ~/environment/echo-backend/src/${version}/
  docker tag echo-backend:${version} ${EKS_ECR_REGISTRY}/echo-backend:${version}
  docker push ${EKS_ECR_REGISTRY}/echo-backend:${version}

  helm -n demos upgrade -i echo-backend-${color} ~/environment/echo-backend/ \
    --set registry=${EKS_ECR_REGISTRY} \
    --set color=${color} \
    --set version=${version}
done
```

Remote into your "jumpbox" and satisfy yourself that the backend services are internally accessible.
```bash
kubectl exec -it jumpbox -- /bin/bash -c "curl http://echo-backend-blue.demos.svc.cluster.local:80; curl http://echo-backend-green.demos.svc.cluster.local:80"
```

In a **dedicated** terminal window, grab the CLB DNS name and put the following `curl` command in a loop.
```bash
clb_dnsname=$(kubectl -n demos get service -l app=echo-frontend-blue -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${clb_dnsname}; sleep 0.25; done
```

In the earlier section titled **Upgrade Your Deployment** we introduced version 2.0 of `echo-frontend` which supports the use of backend apps.
Now is the time to deploy that version into the **blue** environment using an [in-place](https://docs.aws.amazon.com/whitepapers/latest/overview-deployment-options/in-place-deployments.html) strategy.
The driver for reusing the **blue** environment is simply to limit the volume of pods created.

Return to your original terminal window.
As you deploy version 2.0 of your `echo-frontend` app you will tell it where to locate a compatible instance of `echo-backend` as follows.
```bash
helm -n demos upgrade -i echo-frontend-blue ~/environment/echo-frontend/ \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=2.0 \
  --set backend=http://echo-backend-blue.demos.svc.cluster.local:80 \
  --set serviceType=LoadBalancer
```

... TODO
...... get frontend v2.0 to pull out backend.version so it doesn't build embedded JSON
...... start building the app mesh components.

[Return To Main Menu](/README.md)
