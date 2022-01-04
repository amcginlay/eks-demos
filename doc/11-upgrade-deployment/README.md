# Upgrade your deployment

Rapid, iterative code changes are commonplace in cloud native software deployments and Kubernetes copes well with these demands.
You will now make a small change to your code and redeploy your app using `kubectl`.

Version 2.0 of your app provides support for the use of a **backend** app which will be introduced in a later chapter.

Run the following snippet in the terminal to download the new source code and Dockerfile for your app.
```bash
mkdir -p ~/environment/echo-frontend/src/2.0/
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-frontend/src/2.0/main.go \
  -O ~/environment/echo-frontend/src/2.0/main.go
wget https://raw.githubusercontent.com/${EKS_GITHUB_USER}/eks-demos/main/echo-frontend/src/2.0/Dockerfile \
  -O ~/environment/echo-frontend/src/2.0/Dockerfile
```

The target files will be in `~/environment/echo-frontend/src/2.0/`.
Open these files in Cloud9 IDE to review the updated code.
Observe that the Dockerfile is now using a [multi-stage build](https://docs.docker.com/develop/develop-images/multistage-build/).
This will help keep the container image size down to a minimum.

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

Observe the new `backend` attribute ("none" by default) and the value for the `version` attribute which is set to 2.0.

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

You may also like to visit [https://us-west-2.console.aws.amazon.com/ecr/repositories](https://us-west-2.console.aws.amazon.com/ecr/repositories), open up the `echo-frontend` repository and inspect your images via the console.

Re-apply the deployment manifest, adjusting only for the new version, to update your app **in-place**.
```bash
cat ~/environment/echo-frontend/templates/echo-frontend-deployment.yaml | \
    sed "s/{{ .Values.registry }}/${EKS_ECR_REGISTRY}/g" | \
    sed "s/{{ .Values.color }}/blue/g" | \
    sed "s/{{ .Values.version }}/2.0/g" | \
    sed "s/{{ .Values.backend }}/none/g" | \
    kubectl -n demos apply -f -
```

Inspect your updated deployment.
Observe the version change from 1.0 to 2.0 under the "IMAGES" heading.
```bash
sleep 10 && kubectl -n demos get deployments,pods -o wide
```

Exec into the first pod to perform a curl test.
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
    sed "s/{{ .Values.backend }}/none/g" | \
    kubectl -n demos apply -f -
```

The version 2.0 image remains in ECR for later use.

[Return To Main Menu](/README.md)