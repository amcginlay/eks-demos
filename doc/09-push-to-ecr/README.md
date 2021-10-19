# Push Container Image To ECR

Create target ECR repo, deleting it first if needed.
```bash
aws ecr delete-repository --repository-name ${EKS_APP} --force >/dev/null 2>&1
aws ecr create-repository \
  --repository-name ${EKS_APP} \
  --region ${AWS_DEFAULT_REGION} \
  --image-scanning-configuration scanOnPush=true \
  --query 'repository.repositoryUri' \
  --output text
```

NOTE the generated repo name, which is output by the above command, should match our exported variable `EKS_APP_ECR_REPO`.

Push the Docker image to ECR repository.
```bash
aws ecr get-login-password --region ${AWS_DEFAULT_REGION} | docker login --username AWS --password-stdin ${EKS_APP_ECR_REPO}
docker tag ${EKS_APP}:latest ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION}
docker images
docker push ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION}
```

Before we move on, push out the next version of our simple app so we've got something extra to play with.
This might usually involve some real code changes.
In this case we're just incrementing the value of the `VERSION` environment variable inside the container image.

```bash
sed -i "s/ENV VERSION=${EKS_APP_VERSION}/ENV VERSION=${EKS_APP_VERSION_NEXT}/g" ./eks-demos/src/php-echo/Dockerfile
docker build -t ${EKS_APP} ~/environment/eks-demos/src/${EKS_APP_NAME}/
docker tag ${EKS_APP}:latest ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION_NEXT}
docker push ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION_NEXT}
```

The EKS cluster can now locate these images by their version tags.

[Return To Main Menu](/README.md)
