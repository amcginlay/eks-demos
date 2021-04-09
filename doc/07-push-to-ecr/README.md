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

NOTE the generated repo name, which is output by the above command, should match our exported variable `APP_ECR_REPO`.

Push the Docker image to ECR repository.
```bash
aws ecr get-login-password --region ${AWS_DEFAULT_REGION} | docker login --username AWS --password-stdin ${EKS_APP_ECR_REPO}
docker tag ${EKS_APP_NAME}:latest ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION}
docker images
docker push ${EKS_APP_ECR_REPO}:${EKS_APP_VERSION}
```

The EKS cluster can now locate this image by its tag

[Return To Main Menu](/README.md)
