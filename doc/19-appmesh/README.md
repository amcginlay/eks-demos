# (WIP) AWS App Mesh - because managing microservices at scale is hard

This section assumes you have completed the earlier section on **Helm** and that the current `echo-frontend` deployment is under its control.

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

### Deploy bankend services

We aim to use App Mesh to demonstrate a [Blue-green deployment](https://en.wikipedia.org/wiki/Blue-green_deployment) whereby traffic between deployments (i.e. microservices) can be dynamically shifted from one target to another without the need to re-deploy or directly reconfigure any existing workloads.
Using skills introduced in previous chapters you will now rapidly build, containerize, register and deploy two versions of the `echo-backend` app.
```bash
aws ecr delete-repository --repository-name echo-backend --force >/dev/null 2>&1
aws ecr create-repository \
  --repository-name echo-backend \
  --image-scanning-configuration scanOnPush=true \
  --query 'repository.repositoryUri' \
  --output text

aws ecr get-login-password | docker login --username AWS --password-stdin ${EKS_ECR_REGISTRY}

mkdir -p ~/environment/echo-backend/templates/
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

Review the deployed apps, from the perspective of both `helm` and `kubectl`.
```bash
helm -n demos list
kubectl -n demos get deployments,services
```

Remote into your "jumpbox" and satisfy yourself that the backend services are internally accessible.
```bash
kubectl exec -it jumpbox -- /bin/bash -c "curl http://echo-backend-blue.demos.svc.cluster.local:80; curl http://echo-backend-green.demos.svc.cluster.local:80"
```

In a **dedicated** terminal window, grab the CLB DNS name and start making calls to `echo-frontend-blue`.
```bash
clb_dnsname=$(kubectl -n demos get service -l app=echo-frontend-blue -o jsonpath='{.items[0].status.loadBalancer.ingress[0].hostname}')
while true; do curl http://${clb_dnsname}; sleep 0.25; done
```

Leave the **dedicated** terminal window in this state and return to your original terminal window.

In the earlier section titled **Prepare Upgraded Image** we introduced version 2.0 of `echo-frontend` which supports the use of backend apps.
Now is the time to deploy that version using an [in-place](https://docs.aws.amazon.com/whitepapers/latest/overview-deployment-options/in-place-deployments.html) strategy.
The driver for reusing the **blue** environment is simply to limit the volume of pods created.

Whilst you deploy version 2.0 of your `echo-frontend` app you will also provide it with the URL for a compatible instance of `echo-backend` as follows.
Observe that the URL is for the `blue` backend which is acceptable for testing connectivity but creates a tightly coupled relationship between our services.
We instead want out services to be loosely coupled and resolving that issue is the crux of this exercise.
```bash
helm -n demos upgrade -i echo-frontend-blue ~/environment/echo-frontend/ \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=2.0 \
  --set backend=http://echo-backend-blue.demos.svc.cluster.local:80 \
  --set serviceType=LoadBalancer
```

Return to your **dedicated** terminal window to observe the 2.0 `echo-frontend` successfully retrieving `"backend":"1.0"` from the `blue` backend environment.

The next step is to start rollimg out the [AWS App Mesh components](https://docs.aws.amazon.com/app-mesh/latest/userguide/what-is-app-mesh.html#app_mesh_components).
Go the the [App Mesh console](https://us-west-2.console.aws.amazon.com/appmesh/meshes) page.
There is likely to be no Meshes shown here.
Each `Mesh` component encapsulate a logical collection of interconnected service mesh resources.

Download the manifest for your `Mesh` to your Cloud9 environment.
```bash
mkdir -p ~/environment/mesh/templates/
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/mesh/templates/demos-mesh.yaml \
  -O ~/environment/mesh/templates/demos-mesh.yaml
```

Observe the manifest as it gets passed to `kubectl apply`.
```bash
cat ~/environment/mesh/templates/demos-mesh.yaml | \
    tee /dev/tty | \
    kubectl -n demos apply -f -
```

The App Mesh Controller will react by building the new `Mesh` component in the your AWS account so return to the App Mesh console to observe this and click on the mesh named `demos` to reveal that it comprises, among other things, `Virtual nodes` of which there are currently none.
You must resist any temptation to create Mesh components via the console.
Always treat the App Mesh console as a read-only interface.

TODO ....

[Return To Main Menu](/README.md)
