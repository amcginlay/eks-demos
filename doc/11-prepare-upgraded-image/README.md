# Prepare Upgraded Image

Rapid, iterative code changes are commonplace in cloud native software deployments and Kubernetes copes well with these demands.
You will now make a small change to your code, build a new container and push this image to ECR.

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

The version 2.0 image remains in ECR for later use.

[Return To Main Menu](/README.md)