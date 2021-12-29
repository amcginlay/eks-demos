# Upgrade your deployment

Rapid, iterative code changes are commonplace in cloud native software deployments and Kubernetes copes well with these demands.
You will now make a small change to your code and redeploy your app using `kubectl`.

Version 2.0 of your app provides support for the use of a **backend** app which will be introduced in a later chapter.

Run the following snippet in the terminal to pull down the new source code and Dockerfile for your app.
```bash
mkdir -p ~/environment/echo-frontend/src/2.0/
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-frontend/src/2.0/main.go \
     https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-frontend/src/2.0/Dockerfile \
     --directory-prefix ~/environment/echo-frontend/src/2.0/
```

Open `~/environment/echo-frontend/src/2.0/main.go` in Cloud9 IDE to review the updated code.

Use Docker to build and run your new container image.
```bash
docker build -t echo-frontend:2.0 ~/environment/echo-frontend/src/2.0/
container_id=$(docker run --detach --rm -p 8082:80 echo-frontend:2.0)
```

Give it a quick test then stop the running container.
```bash
curl http://localhost:8082
docker stop ${container_id}
```

Observe the new `backend` attribute ("n/a" by default) and the value for the `version` attribute which is set to 2.0.

Tag and push the Docker image to the ECR repository.
```bash
docker tag echo-frontend:2.0 ${EKS_ECR_REGISTRY}/echo-frontend:2.0
aws ecr get-login-password | docker login --username AWS --password-stdin ${EKS_ECR_REGISTRY}
docker push ${EKS_ECR_REGISTRY}/echo-frontend:2.0
```

Review the version 1.0 and version 2.0 images, now side by side in ECR.
```bash
aws ecr list-images --repository-name echo-frontend
```

Re-apply the deployment manifest, adjusting only for the new version, to update your app **in-place**.
```bash
cat ~/environment/echo-frontend/templates/echo-frontend-deployment.yaml | \
    sed "s/{{ .Values.registry }}/${EKS_ECR_REGISTRY}/g" | \
    sed "s/{{ .Values.color }}/blue/g" | \
    sed "s/{{ .Values.version }}/2.0/g" | \
    kubectl apply -f -
```

Inspect your updated deployment.
Observe the version change from 1.0 to 2.0 under the "IMAGES" heading.
```bash
sleep 10 && kubectl -n demos get deployments,pods -o wide
```

Exec into the first pod to perform curl test.
Satisfy yourself that your app has been upgraded.
```bash
first_pod=$(kubectl -n demos get pods -l app=echo-frontend-blue -o name | head -1)
kubectl -n demos exec -it ${first_pod} -- curl http://localhost:80
```

For now, roll back your deployment to version 1.0.
```bash
cat ~/environment/echo-frontend/templates/echo-frontend-deployment.yaml | \
    sed "s/{{ .Values.registry }}/${EKS_ECR_REGISTRY}/g" | \
    sed "s/{{ .Values.color }}/blue/g" | \
    sed "s/{{ .Values.version }}/1.0/g" | \
    kubectl apply -f -
```

The version 2.0 image remains in ECR for later use.

[Return To Main Menu](/README.md)