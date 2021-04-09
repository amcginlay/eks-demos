# Push Container Image To ECR

Create target ECR repo, deleting it first if needed.
```bash
aws ecr delete-repository --repository-name ${EKS_APP_NAME} --force >/dev/null 2>&1
aws ecr create-repository \
  --repository-name ${EKS_APP_NAME} \
  --region ${AWS_DEFAULT_REGION} \
  --image-scanning-configuration scanOnPush=true \
  --query 'repository.repositoryUri' \
  --output text
```

NOTE the generated repo name output by the above command should match our exported variable APP_ECR_REPO.

Push the Docker image to ECR repository.
```bash
aws ecr get-login-password --region ${AWS_DEFAULT_REGION} | docker login --username AWS --password-stdin ${APP_ECR_REPO}
docker tag ${APP_NAME}:latest ${APP_ECR_REPO}:${APP_VERSION}
docker images
docker push ${APP_ECR_REPO}:${APP_VERSION}
```

The EKS cluster can now locate this image by its tag

[Return To Main Menu](/README.md)
