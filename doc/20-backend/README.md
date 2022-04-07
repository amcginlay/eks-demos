# Deploy Backend Services - because no one likes a monolith

This section assumes the following:
- you have completed the earlier section on **Helm**
- the current `echo-frontend` deployment is version **v2.0** and is deployed under the control of Helm.
- You have an `nginx` "jumpbox" installed as described in **"K8s ClusterIP Services"**

We aim to use App Mesh to demonstrate a [Blue-green deployment](https://en.wikipedia.org/wiki/Blue-green_deployment) whereby, once stood up, the traffic between deployments (i.e. microservices) can be dynamically shifted from one target to another without the need to re-deploy or directly reconfigure **any** existing workloads.
There are many other benefits of implementing a service mesh but this is the prominent one which will use in this demo.

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

# next two looping sections require these version mappings
declare -A versions=()
versions[blue]=11.0
versions[green]=12.0

# BLUE and GREEN backends are built and labeled v11.0 and v12.0 respectively
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
done

# Deploy the BLUE and GREEN backends
for color in blue green; do
  version=${versions[${color}]}
  helm -n demos upgrade -i echo-backend-${color} ~/environment/echo-backend/ \
    --create-namespace \
    --set registry=${EKS_ECR_REGISTRY} \
    --set color=${color} \
    --set version=${version}
done
```

Review the deployed apps, from the perspective of `ecr`, `helm` and `kubectl`.
```bash
aws ecr list-images --repository-name echo-backend
helm -n demos list
kubectl -n demos get deployments,services -o wide
```

Observe how the `blue` and `green` backend deployments each have a **single** replica pod which utilize different images.

Remote into your "jumpbox" and satisfy yourself that the **backend** services are internally accessible.
```bash
kubectl exec -it jumpbox -- /bin/bash -c "curl http://echo-backend-blue.demos.svc.cluster.local:80; curl http://echo-backend-green.demos.svc.cluster.local:80"
```

In a **dedicated** terminal window, run a similar looped command against the **frontend**, which you will note is **not** currently configured to use any **backend**.
```bash
kubectl exec -it jumpbox -- /bin/bash -c "while true; do curl http://echo-frontend-blue.demos.svc.cluster.local:80; sleep 0.25; done"
```

Leave the **dedicated** terminal window in this state and return to your original terminal window.

Now **redeploy** your `echo-frontend` app, this time providing the URL for a compatible instance of `echo-backend` as follows.
```bash
helm -n demos upgrade -i echo-frontend-blue ~/environment/echo-frontend/ \
  --create-namespace \
  --set registry=${EKS_ECR_REGISTRY} \
  --set color=blue \
  --set version=2.0 \
  --set backend=http://echo-backend-blue.demos.svc.cluster.local:80 \
  --set serviceType=ClusterIP
```

Return to your **dedicated** terminal window to observe the frontend successfully retrieving `"backend":"11.0"` from the `blue` backend URL.

As an exercise you can attempt yourself, try redeploying the frontend with the backend URL pointing at the `green` backend.

If something feels wrong about this configuration you'd be correct.
If you choose to implement this blue/green pattern for **all** future backend rollouts, you cannot avoid the need to reconfigure and redeploy the frontend **every time** the backend is changed.
This introduces **unintended risk** to your deployments as well as **tight coupling** between the two deployments.
Instead, you want your deployments to be **risk averse** and **loosely coupled**.
A solution to this problem is presented in the next chapter on **App Mesh**.

[Return To Main Menu](/README.md)
