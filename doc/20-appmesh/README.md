# AWS App Mesh - because managing microservices at scale is hard

This section assumes the following:
- you have completed the earlier section on **Helm**
- the current `echo-frontend` deployment is under control of Helm.
- additionally the `echo-frontend` deployment version is **v2.0**

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
  --set serviceAccount.name=appmesh-controller
```

Verify that the App Mesh Controller is installed.
```bash
kubectl -n kube-system get deployment appmesh-controller
```

### Deploy backend services

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

In the earlier section titled **Prepare Upgraded Image** you created version 2.0 of `echo-frontend` which supports the use of backend apps.
Now is the time to deploy that version using an [in-place](https://docs.aws.amazon.com/whitepapers/latest/overview-deployment-options/in-place-deployments.html) strategy.
The driver for reusing the **blue** environment is simply to limit the volume of pods created.

Whilst you deploy version 2.0 of your `echo-frontend` app you will also provide it with the URL for a compatible instance of `echo-backend` as follows.
Observe that the URL is for the `blue` backend which is acceptable for testing connectivity but creates a tightly coupled relationship between your services.
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

The next step is to start rolling out the [AWS App Mesh components](https://docs.aws.amazon.com/app-mesh/latest/userguide/what-is-app-mesh.html#app_mesh_components).
Go the the [App Mesh console](https://us-west-2.console.aws.amazon.com/appmesh/meshes) page.
There is likely to be no Meshes currently shown here.
Each `Mesh` resource encapsulates a logical collection of other interconnected service mesh resources.

Download the manifest for your `Mesh` component to your Cloud9 environment.
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

As the `Mesh` object is introduced to the Kubernetes cluster the App Mesh Controller reacts by building the new AppMesh **Mesh** resource in your AWS account.
Use the following commands to view these **twinned** resources.
```bash
#K8s
kubectl -n demos get meshes
kubectl -n demos describe mesh demos
#AWS
aws appmesh list-meshes
aws appmesh describe-mesh --mesh-name demos
```

Return to the App Mesh console to observe this and click on the mesh named `demos` to reveal that it comprises, among other things, `Virtual nodes` of which there are currently none.
You must resist any temptation to create Mesh components via the console.
Always treat the App Mesh console as a read-only interface.

Download the manifests for your `VirtualNode` **backend** components to your Cloud9 environment.
```bash
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/mesh/templates/vn-echo-backend-blue.yaml \
  -O ~/environment/mesh/templates/vn-echo-backend-blue.yaml
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/mesh/templates/vn-echo-backend-green.yaml \
  -O ~/environment/mesh/templates/vn-echo-backend-green.yaml
```

Observe the manifests as they gets passed to `kubectl apply`.
```bash
cat ~/environment/mesh/templates/vn-echo-backend-blue.yaml \
    <(echo "---") \
    ~/environment/mesh/templates/vn-echo-backend-green.yaml | \
  tee /dev/tty | \
  kubectl -n demos apply -f -
```

As the `VirtualNode` **backend** objects are introduced to the Kubernetes cluster the App Mesh Controller reacts by adding them to the your **demos** mesh.
Use the following commands to view these resources.
```bash
#K8s
kubectl -n demos get virtualnodes
kubectl -n demos describe virtualnode vn-echo-backend-blue
kubectl -n demos describe virtualnode vn-echo-backend-green
#AWS
aws appmesh list-virtual-nodes --mesh-name demos
aws appmesh describe-virtual-node --mesh-name demos \
  --virtual-node-name vn-echo-backend-blue
aws appmesh describe-virtual-node --mesh-name demos \
  --virtual-node-name vn-echo-backend-green
```

Note that the **frontend** `VirtualNode` resource is deferred until later as it depends upon the backend services being fully "meshified".

**TODO** is this 100% true?
We must first introduce a `VirtualService` and (in this case) a `VirtualRouter` resource to marshal requests through to the backends.

Download the manifests for your `VirtualRouter` component to your Cloud9 environment.
```bash
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/mesh/templates/vr-echo-backend.yaml \
  -O ~/environment/mesh/templates/vr-echo-backend.yaml
```

Observe the manifests as it gets passed to `kubectl apply`.
```bash
cat ~/environment/mesh/templates/vr-echo-backend.yaml | \
  tee /dev/tty | \
  kubectl -n demos apply -f -
```

As the `VirtualRouter` object and its routes are introduced to the Kubernetes cluster the App Mesh Controller reacts by adding them to the your **demos** mesh.
Use the following commands to view these resources.
```bash
#K8s
kubectl -n demos get virtualrouters
#AWS
aws appmesh list-virtual-routers --mesh-name demos
aws appmesh list-routes --mesh-name demos \
  --virtual-router-name vr-echo-backend
aws appmesh describe-route --mesh-name demos \
  --virtual-router-name vr-echo-backend \
  --route-name vrr-echo-backend
```

Inspect the output carefully.
You will observe that the route, named `vrr-echo-backend`, is **weighted** to forward 100% of incoming traffic to `vn-echo-backend-blue`.
This is intentional, and it's a topic we will return to.

Your next step is to introduce a `VirtualService` object which depends upon the `VirtualRouter` object you just created as well as an underlying Kubernetes service object.
The Kubernetes service you twin with your `VirtualService` doesn't need to resolve to any endpoints/pods, it just needs to surface a cluster IP address for identity purposes.

Download the manifests for your `VirtualService` and its associated `Service` component to your Cloud9 environment.
```bash
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/mesh/templates/vs-echo-backend.yaml \
  -O ~/environment/mesh/templates/vs-echo-backend.yaml
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/mesh/templates/vs-echo-backend-service.yaml \
  -O ~/environment/mesh/templates/vs-echo-backend-service.yaml
```

Observe the manifests as they gets passed to `kubectl apply`.
```bash
cat ~/environment/mesh/templates/vs-echo-backend.yaml \
    <(echo "---") \
    ~/environment/mesh/templates/vs-echo-backend-service.yaml | \
  tee /dev/tty | \
  kubectl -n demos apply -f -
```

As the `VirtualService` object is introduced to the Kubernetes cluster the App Mesh Controller reacts by adding it to the your **demos** mesh.
The associated `Service` objects is created as normal but, as the `spec:selector:` section is missing from its manifest, it has no associated endpoints (i.e. targets).
Use the following commands to view these resources.
```bash
#K8s
kubectl -n demos get virtualservices
kubectl -n demos get services
kubectl -n demos get endpoints
#AWS
aws appmesh list-virtual-services --mesh-name demos
aws appmesh list-routes --mesh-name demos \
  --virtual-router-name vr-echo-backend
aws appmesh describe-route --mesh-name demos \
  --virtual-router-name vr-echo-backend \
  --route-name vrr-echo-backend
```

As the `VirtualService` resource it depends upon is now in place, we can now create the deferred **frontend** `VirtualNode` resource.

Download the manifest for your `VirtualNode` **frontend** component to your Cloud9 environment.
```bash
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/mesh/templates/vn-echo-frontend-blue.yaml \
  -O ~/environment/mesh/templates/vn-echo-frontend-blue.yaml
```

Observe the manifests as they gets passed to `kubectl apply`.
```bash
cat ~/environment/mesh/templates/vn-echo-frontend-blue.yaml | \
  tee /dev/tty | \
  kubectl -n demos apply -f -
```

As the `VirtualNode` **frontend** object is introduced to the Kubernetes cluster the App Mesh Controller reacts by adding it to the your **demos** mesh.
Use the following commands to view these resources.
```bash
#K8s
kubectl -n demos get virtualnodes
kubectl -n demos describe virtualnode vn-echo-frontend
#AWS
aws appmesh list-virtual-nodes --mesh-name demos
aws appmesh describe-virtual-node --mesh-name demos \
  --virtual-node-name vn-echo-frontend
```

For workloads to support AppMesh they require the injection of the [envoy](https://www.envoyproxy.io/) service proxy which runs as a sidecar container.
Start by activating your namespace for use with your new mesh by applying a pair of labels as follows.
```bash
kubectl label namespace demos mesh=demos
kubectl label namespace demos appmesh.k8s.aws/sidecarInjectorWebhook=enabled
```

Check the labels have been applied.
```bash
kubectl describe namespace demos
```

Now restart all your deployments to get envoy injected
```bash
kubectl -n demos rollout restart deployment \
  echo-frontend-blue \
  echo-backend-blue \
  echo-backend-green
```

There is currently no change since `echo-frontend-blue` is still configured to point at `echo-backend-blue` (i.e. v1.0).
We could, of course, reconfigure the frontend to point at `echo-backend-green` (i.e. v2.0) but that means having to reconfigure the the frontend **every time** the backend is updated.
Instead, by pointing frontend at `vs-echo-backend` which is configured via AppMesh, we benefit from externalizing that configuration from now on.

Reconfigure your frontend to target the virtualized service which is now discoverable, via the envoy proxy service, at `vs-echo-backend`.
```bash
helm -n demos upgrade -i echo-frontend-blue ~/environment/echo-frontend/ \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=2.0 \
  --set backend=http://vs-echo-backend:80 \
  --set serviceType=LoadBalancer
```

Return to your **dedicated** terminal window which is making calls to the CLB DNS name.
You may recall that the backend weights were 100% blue and 0% green so, thus far, it looks like nothing has changed in the behavior of your app.
Let's modify the weights in the `VirtualRouter` to split the traffic more evenly.
```bash
sed -i "s/weight: 100/weight: 50/g" ~/environment/mesh/templates/vr-echo-backend.yaml
sed -i "s/weight: 0/weight: 50/g" ~/environment/mesh/templates/vr-echo-backend.yaml
kubectl -n demos apply -f ~/environment/mesh/templates/vr-echo-backend.yaml
```

[Return To Main Menu](/README.md)
